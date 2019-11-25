package release

import (
	"context"
	"fmt"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/operatorkit/resource/crud"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	releaseState, err := toReleaseState(updateChange)
	if err != nil {
		return microerror.Mask(err)
	}

	upgradeForce := key.HasForceUpgradeAnnotation(cr)

	if releaseState.Name != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating release %#q with force == %t", releaseState.Name, upgradeForce))

		tarballURL := key.TarballURL(cr)
		tarballPath, err := r.helmClient.PullChartTarball(ctx, tarballURL)
		if helmclient.IsPullChartFailedError(err) {
			r.logger.LogCtx(ctx, "level", "warning", "message", "pulling chart failed", "stack", microerror.Stack(err))

			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		defer func() {
			err := r.fs.Remove(tarballPath)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %#q failed", tarballPath), "stack", fmt.Sprintf("%#v", err))
			}
		}()

		// We need to pass the ValueOverrides option to make the update process
		// use the default values and prevent errors on nested values.
		err = r.helmClient.UpdateReleaseFromTarball(ctx, releaseState.Name, tarballPath,
			helm.UpdateValueOverrides(releaseState.ValuesYAML),
			helm.UpgradeForce(upgradeForce))
		if err != nil {
			releaseContent, err := r.helmClient.GetReleaseContent(ctx, releaseState.Name)
			if helmclient.IsReleaseNotFound(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("helm release not found %#q", releaseContent.Name))
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)
				return nil
			} else if err != nil {
				return microerror.Mask(err)
			}
			if releaseContent.Status == helmFailedStatus {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("failed to update release %#q", releaseContent.Name))
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)
				return nil
			}
			return microerror.Mask(err)
		}

		err = r.patchAnnotations(ctx, cr, releaseState)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated release %#q", releaseState.Name))
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	create, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	update, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetCreateChange(create)
	patch.SetUpdateChange(update)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentReleaseState, err := toReleaseState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredReleaseState, err := toReleaseState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if the %#q release has to be updated", desiredReleaseState.Name))

	if isReleaseInTransitionState(currentReleaseState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release is in status %#q and cannot be updated", desiredReleaseState.Name, currentReleaseState.Status))
		return nil, nil
	}

	isModified := isReleaseModified(currentReleaseState, desiredReleaseState)
	isWrongStatus := isWrongStatus(currentReleaseState, desiredReleaseState)

	if isModified || isWrongStatus {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release has to be updated", desiredReleaseState.Name))

		return &desiredReleaseState, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release does not have to be updated", desiredReleaseState.Name))
	}

	return nil, nil
}
