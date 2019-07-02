package chartstatus

import (
	"context"
	"fmt"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"

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
			} else {
				reason = ""
			}
		}
	}

	if key.ReleaseStatus(customObject) != status {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting status for release '%s' status to '%s'", releaseName, status))

		customObjectCopy := customObject.DeepCopy()
		customObjectCopy.Status.ReleaseStatus = status
		customObjectCopy.Status.Reason = reason

		_, err := r.g8sClient.CoreV1alpha1().ChartConfigs(customObject.Namespace).UpdateStatus(customObjectCopy)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status set for release '%s'", releaseName))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status for release '%s' already set to '%s'", releaseName, status))
	}

	return nil
}
