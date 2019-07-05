package chartstatus

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
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
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("getting status for release '%s'", releaseName))

	releaseContent, err := r.helmClient.GetReleaseContent(ctx, releaseName)
	if helmclient.IsReleaseNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release '%s' not found", releaseName))

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

	chartConfigStatus := v1alpha1.ChartConfigStatus{
		ReleaseStatus: status,
		Reason:        reason,
	}

	if customObject.Status != chartConfigStatus {
		// Get chartconfig CR again to ensure the resource version is correct.
		currentCR, err := r.g8sClient.CoreV1alpha1().ChartConfigs(customObject.Namespace).Get(customObject.Name, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		if currentCR.Status.Reason != chartConfigStatus.Reason {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting reason for release '%s' to '%s'", releaseName, reason))
			currentCR.Status.Reason = chartConfigStatus.Reason
		}

		if currentCR.Status.ReleaseStatus != chartConfigStatus.ReleaseStatus {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting status for release '%s' status to '%s'", releaseName, status))
			currentCR.Status.ReleaseStatus = chartConfigStatus.ReleaseStatus
		}

		_, err = r.g8sClient.CoreV1alpha1().ChartConfigs(customObject.Namespace).UpdateStatus(currentCR)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status set for release '%s'", releaseName))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status for release %#q already set to %#q", releaseName, status))
	}

	return nil
}
