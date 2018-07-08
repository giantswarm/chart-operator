package chartstatus

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/service/controller/v2/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	releaseName := key.ReleaseName(customObject)
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("getting status for release '%s'", releaseName))

	releaseContent, err := r.helmClient.GetReleaseContent(releaseName)
	// TODO
	if err != nil {
		return microerror.Mask(err)
	}

	if key.ReleaseStatus(customObject) != releaseContent.Status {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("setting status for release '%s' status to '%s'", releaseName, releaseContent.Status))

		customObjectCopy := customObject.DeepCopy()
		customObjectCopy.Status.ReleaseStatus = releaseContent.Status

		_, err := r.g8sClient.CoreV1alpha1().ChartConfigs(customObject.Namespace).UpdateStatus(customObjectCopy)
		// TODO
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status set for release '%s'", releaseName))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("status for release '%s' already set to '%s'", releaseName, releaseContent.Status))
	}

	return nil
}
