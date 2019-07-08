package chartstatus

import (
	"context"
	"fmt"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v7/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	releaseName := key.ReleaseName(customObject)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("getting status for release %#q", releaseName))

	releaseContent, err := r.helmClient.GetReleaseContent(ctx, releaseName)
	if helmclient.IsReleaseNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q not found", releaseName))

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		// Return early. We will retry on the next execution.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	var status, reason string
	{
		if key.IsCordoned(customObject) {
			status = releaseStatusCordoned
			reason = key.CordonReason(customObject)
		} else {
			status = releaseContent.Status
			if releaseContent.Status != releaseStatusDeployed {
				releaseHistory, err := r.helmClient.GetReleaseHistory(ctx, releaseName)
				if err != nil {
					return microerror.Mask(err)
				}

				reason = releaseHistory.Description
			}
		}
	}

	if customObject.Status.ReleaseStatus != status || customObject.Status.Reason != reason {
		// Get chartconfig CR again to ensure the resource version is correct.
		currentCR, err := r.g8sClient.CoreV1alpha1().ChartConfigs(customObject.Namespace).Get(customObject.Name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		if currentCR.Status.Reason != reason {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting reason for release %#q to %#q", releaseName, reason))
			currentCR.Status.Reason = reason
		}

		if currentCR.Status.ReleaseStatus != status {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting status for release %#q status to %#q", releaseName, status))
			currentCR.Status.ReleaseStatus = status
		}

		_, err = r.g8sClient.CoreV1alpha1().ChartConfigs(customObject.Namespace).UpdateStatus(currentCR)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status set for release %#q", releaseName))
	}

	return nil
}
