package release

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/operatorkit/resource/crud"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/chart-operator/pkg/annotation"
	"github.com/giantswarm/chart-operator/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/service/controller/chart/key"
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

	if releaseState.Name == "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("no release name is provided for %#q", cr.Name))
		return nil
	}

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

	// TODO: Disabling helm upgrade --force from chart-operator since recreate
	// is not supported.
	//
	//	See https://github.com/giantswarm/giantswarm/issues/11376
	//
	upgradeForce := key.HasForceUpgradeAnnotation(cr)
	if upgradeForce {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("helm upgrade force is disabled for %#q", releaseState.Name))
	}

	ch := make(chan error)

	// We update the helm release but with a wait timeout so we don't
	// block reconciling other CRs.
	//
	// If we do timeout the update will continue in the background.
	// We will check the progress in the next reconciliation loop.
	go func() {
		opts := helmclient.UpdateOptions{
			Force: false,
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
	case <-time.After(r.k8sWaitTimeout):
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for %d secs. release still being updated", int64(r.k8sWaitTimeout.Seconds())))

		// The update will continue in the background. We set the checksum
		// annotation so the update state calculation is accurate when we check
		// in the next reconciliation loop.
		err = r.patchAnnotations(ctx, cr, releaseState)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	if helmclient.IsValidationFailedError(err) {
		reason := err.Error()
		reason = fmt.Sprintf("helm validation error: (%s)", reason)
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("helm release %#q failed, %s", releaseState.Name, reason))
		addStatusToContext(cc, reason, validationFailedStatus)

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if helmclient.IsInvalidManifest(err) {
		reason := err.Error()
		reason = fmt.Sprintf("invalid manifest error: (%s)", reason)
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("helm release %#q failed, %s", releaseState.Name, reason))
		addStatusToContext(cc, reason, invalidManifestStatus)

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if err != nil {
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

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated release %#q", releaseState.Name))

	// We set the checksum annotation so the update state calculation
	// is accurate when we check in the next reconciliation loop.
	err = r.patchAnnotations(ctx, cr, releaseState)
	if err != nil {
		return microerror.Mask(err)
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
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	currentReleaseState, err := toReleaseState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredReleaseState, err := toReleaseState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if the %#q release has to be updated", desiredReleaseState.Name))

	counts, rollbackCountsExist := cr.GetAnnotations()[annotation.RollbackCounts]
	var rollbackCounts int
	if !rollbackCountsExist {
		rollbackCounts = 0
	} else {
		rollbackCounts, err = strconv.Atoi(counts)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// The release is still being updated so we don't update and check again
	// in the next reconciliation loop.
	if isReleaseInTransitionState(currentReleaseState) {
		if rollbackCounts > maxRollBack {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release is in status %#q and cannot be updated", desiredReleaseState.Name, currentReleaseState.Status))
			return nil, nil
		}

		if currentReleaseState.Status == helmclient.StatusPendingInstall {
			err = r.helmClient.DeleteRelease(ctx, key.Namespace(cr), key.ReleaseName(cr))
			if err != nil {
				return nil, microerror.Mask(err)
			}
		} else {
			// Rollback to revision 0 restore a release to the previous revision.
			err = r.helmClient.Rollback(ctx, key.Namespace(cr), key.ReleaseName(cr), 0, helmclient.RollbackOptions{})
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		patches := []Patch{
			{
				Op:    "add",
				Path:  fmt.Sprintf("/metadata/annotations/%s", annotation.RollbackCounts),
				Value: fmt.Sprintf("%d", rollbackCounts+1),
			},
		}

		bytes, err := json.Marshal(patches)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		_, err = r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).Patch(cr.Name, types.JSONPatchType, bytes)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		return nil, nil
	}

	if rollbackCountsExist {
		patches := []Patch{
			{
				Op:   "remove",
				Path: fmt.Sprintf("/metadata/annotations/%s", annotation.RollbackCounts),
			},
		}

		bytes, err := json.Marshal(patches)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		_, err = r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).Patch(cr.Name, types.JSONPatchType, bytes)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// The release is failed and the values and version have not changed. So we
	// don't update. We will be alerted so we can investigate.
	if isReleaseFailed(currentReleaseState, desiredReleaseState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release is in status %#q and cannot be updated", desiredReleaseState.Name, currentReleaseState.Status))
		return nil, nil
	}

	if isReleaseModified(currentReleaseState, desiredReleaseState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release has to be updated", desiredReleaseState.Name))
		return &desiredReleaseState, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q release does not have to be updated", desiredReleaseState.Name))
	}

	return nil, nil
}
