package release

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/operatorkit/resource/crud"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/controllercontext"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
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
			reason := fmt.Sprintf("pulling chart %#q failed", tarballURL)
			addStatusToContext(cc, reason, releaseNotInstalledStatus)

			r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		} else if helmclient.IsPullChartNotFound(err) {
			reason := fmt.Sprintf("chart %#q not found", tarballURL)
			addStatusToContext(cc, reason, releaseNotInstalledStatus)

			r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		} else if helmclient.IsPullChartTimeout(err) {
			reason := fmt.Sprintf("timeout pulling %#q", tarballURL)
			addStatusToContext(cc, reason, releaseNotInstalledStatus)

			r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
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

		ch := make(chan error)

		// We update the helm release but with a short timeout so we don't
		// block reconciling other CRs. This gives time to make the port
		// forwarding connection to the Tiller API.
		//
		// If we do timeout the update will continue in the background.
		// We will check the progress in the next reconciliation loop.
		go func() {
			opts := helmclient.UpdateOptions{
				Force: upgradeForce,
			}

			// We need to pass the ValueOverrides option to make the update process
			// use the default values and prevent errors on nested values.
			err = r.helmClient.UpdateReleaseFromTarball(ctx,
				tarballPath,
				key.Namespace(cr),
				releaseState.Name,
				releaseState.Values,
				opts)
			close(ch)
		}()

		select {
		case <-ch:
			// Fall through.
		case <-time.After(3 * time.Second):
			r.logger.LogCtx(ctx, "level", "debug", "message", "release still being updated")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			return nil
		}

		if err != nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("helm release %#q failed", releaseState.Name), "stack", microerror.JSON(err))

			releaseContent, err := r.helmClient.GetReleaseContent(ctx, key.Namespace(cr), releaseState.Name)
			if helmclient.IsReleaseNotFound(err) {
				reason := fmt.Sprintf("release %#q not found", releaseState.Name)
				addStatusToContext(cc, reason, releaseNotInstalledStatus)

				r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
				resourcecanceledcontext.SetCanceled(ctx)
				return nil

			} else if err != nil {
				return microerror.Mask(err)
			}
			// Release is failed so the status resource will check the Helm release.
			if releaseContent.Status == helmclient.StatusFailed {
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
