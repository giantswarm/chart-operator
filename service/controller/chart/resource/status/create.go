package status

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/service/controller/chart/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	releaseName := key.ReleaseName(cr)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("getting status for release %#q", releaseName))

	releaseContent, err := r.helmClient.GetReleaseContent(ctx, releaseName)
	if helmclient.IsReleaseNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q not found", releaseName))

		// There is no Helm release for this chart CR so its likely that
		// something has gone wrong. This could be for a reason outside
		// of Helm like the tarball URL is incorrect.
		//
		// If something goes wrong outside of Helm we add that to the
		// controller context in the release resource. So we include this
		// information in the CR status.
		if cc.Status.Reason != "" {
			status := v1alpha1.ChartStatus{
				Reason: cc.Status.Reason,
				Release: v1alpha1.ChartStatusRelease{
					Status: cc.Status.Release.Status,
				},
			}

			err = r.setStatus(ctx, cr, status)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		// Return early. We will retry on the next execution.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	releaseHistory, err := r.helmClient.GetReleaseHistory(ctx, releaseName)
	if helmclient.IsReleaseNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not get status for release %#q", releaseName))
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q not found", releaseName))

		// Return early. We will retry on the next execution.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	var status, reason string
	{
		if key.IsCordoned(cr) {
			status = releaseStatusCordoned
			reason = key.CordonReason(cr)
		} else if cc.Status.Reason != "" {
			// A Helm release exists for this chart CR but data was also added
			// to the controller context. We do this if something goes wrong
			// outside of Helm such as pulling the chart tarball. So we ensure
			// both messages are included in the CR status.
			reason = fmt.Sprintf("Helm reason: %s Operator reason: %s", releaseHistory.Description, cc.Status.Reason)
		} else {
			status = releaseContent.Status
			if releaseContent.Status != releaseStatusDeployed {
				reason = releaseHistory.Description
			}
		}
	}

	desiredStatus := v1alpha1.ChartStatus{
		AppVersion: releaseHistory.AppVersion,
		Reason:     reason,
		Release: v1alpha1.ChartStatusRelease{
			LastDeployed: metav1.Time{releaseHistory.LastDeployed},
			Revision:     releaseHistory.Revision,
			Status:       status,
		},
		Version: releaseHistory.Version,
	}

	if !equals(desiredStatus, key.ChartStatus(cr)) {
		err = r.setStatus(ctx, cr, desiredStatus)
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status for release %#q already set to %#q", releaseName, releaseContent.Status))
	}

	return nil
}

func (r *Resource) setStatus(ctx context.Context, cr v1alpha1.Chart, status v1alpha1.ChartStatus) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting status for release %#q status to %#q", key.ReleaseName(cr), status.Release.Status))

	// Get chart CR again to ensure the resource version is correct.
	currentCR, err := r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).Get(cr.Name, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	currentCR.Status = status

	_, err = r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).UpdateStatus(currentCR)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("set status for release %#q", key.ReleaseName(cr)))

	return nil
}
