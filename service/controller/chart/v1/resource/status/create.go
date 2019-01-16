package status

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customResource, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	releaseName := key.ReleaseName(customResource)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("getting status for release %#q", releaseName))

	releaseContent, err := r.helmClient.GetReleaseContent(ctx, releaseName)
	if helmclient.IsReleaseNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q not found", releaseName))

		// Return early. We will retry on the next execution.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	releaseHistory, err := r.helmClient.GetReleaseHistory(ctx, releaseName)
	if helmclient.IsReleaseNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q not found", releaseName))

		// Return early. We will retry on the next execution.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	currentStatus := v1alpha1.ChartStatus{
		AppVersion:  releaseHistory.AppVersion,
		Status:      releaseContent.Status,
		LastUpdated: releaseHistory.LastUpdated,
		Version:     releaseHistory.Version,
	}

	if !Equals(currentStatus, key.ChartStatus(customResource)) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting status for release %#q status to %#q", releaseName, releaseContent.Status))

		customResourceCopy := customResource.DeepCopy()
		customResourceCopy.Status = currentStatus
		_, err := r.g8sClient.CoreV1alpha1().ChartConfigs(customResource.Namespace).UpdateStatus(customResourceCopy)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status set for release %#q", releaseName))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status for release %#q already set to %#q", releaseName, releaseContent.Status))
	}

	return nil
}
