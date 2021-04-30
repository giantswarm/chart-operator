package release

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/operatorkit/v4/pkg/resource/crud"
	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/chart-operator/v2/pkg/annotation"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
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
		// no-op
		return nil
	}

	r.logger.Debugf(ctx, "updating release %#q in namespace %#q", releaseState.Name, key.Namespace(cr))

	tarballURL := key.TarballURL(cr)
	tarballPath, err := r.helmClient.PullChartTarball(ctx, tarballURL)
	if helmclient.IsPullChartFailedError(err) {
		reason := fmt.Sprintf("pulling chart %#q failed", tarballURL)
		addStatusToContext(cc, reason, releaseNotInstalledStatus)

		r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if helmclient.IsPullChartNotFound(err) {
		reason := fmt.Sprintf("chart %#q not found", tarballURL)
		addStatusToContext(cc, reason, releaseNotInstalledStatus)

		r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if helmclient.IsPullChartTimeout(err) {
		reason := fmt.Sprintf("timeout pulling %#q", tarballURL)
		addStatusToContext(cc, reason, releaseNotInstalledStatus)

		r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	defer func() {
		err := r.fs.Remove(tarballPath)
		if err != nil {
			r.logger.Errorf(ctx, err, "deletion of %#q failed", tarballPath)
		}
	}()

	// TODO: Disabling helm upgrade --force from chart-operator since recreate
	// is not supported.
	//
	//	See https://github.com/giantswarm/giantswarm/issues/11376
	//
	upgradeForce := key.HasForceUpgradeAnnotation(cr)
	if upgradeForce {
		r.logger.Debugf(ctx, "helm upgrade force is disabled for %#q", releaseState.Name)
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
		r.logger.Debugf(ctx, "waited for %d secs. release still being updated", int64(r.k8sWaitTimeout.Seconds()))

		// The update will continue in the background. We set the checksum
		// annotation so the update state calculation is accurate when we check
		// in the next reconciliation loop.
		err = r.addHashAnnotation(ctx, cr, releaseState)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	if helmclient.IsResourceAlreadyExists(err) {
		reason := err.Error()
		reason = fmt.Sprintf("resource already exists: (%s)", reason)
		r.logger.Debugf(ctx, "helm release %#q failed, %s", releaseState.Name, reason)
		addStatusToContext(cc, reason, alreadyExistsStatus)

		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if helmclient.IsValidationFailedError(err) {
		reason := err.Error()
		reason = fmt.Sprintf("helm validation error: (%s)", reason)
		r.logger.Debugf(ctx, "helm release %#q failed, %s", releaseState.Name, reason)
		addStatusToContext(cc, reason, validationFailedStatus)

		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if helmclient.IsInvalidManifest(err) {
		reason := err.Error()
		reason = fmt.Sprintf("invalid manifest error: (%s)", reason)
		r.logger.Debugf(ctx, "helm release %#q failed, %s", releaseState.Name, reason)
		addStatusToContext(cc, reason, invalidManifestStatus)

		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if err != nil {
		r.logger.Errorf(ctx, err, "helm release %#q failed", releaseState.Name)

		releaseContent, err := r.helmClient.GetReleaseContent(ctx, key.Namespace(cr), releaseState.Name)
		if helmclient.IsReleaseNotFound(err) {
			reason := fmt.Sprintf("release %#q not found", releaseState.Name)
			addStatusToContext(cc, reason, releaseNotInstalledStatus)

			r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
			r.logger.Debugf(ctx, "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil

		} else if err != nil {
			return microerror.Mask(err)
		}
		// Release is failed so the status resource will check the Helm release.
		if releaseContent.Status == helmclient.StatusFailed {
			addStatusToContext(cc, releaseContent.Description, helmclient.StatusFailed)

			r.logger.Debugf(ctx, "failed to update release %#q", releaseContent.Name)
			r.logger.Debugf(ctx, "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		}
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "updated release %#q in namespace %#q", releaseState.Name, key.Namespace(cr))

	// We set the checksum annotation so the update state calculation
	// is accurate when we check in the next reconciliation loop.
	err = r.addHashAnnotation(ctx, cr, releaseState)
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

	r.logger.Debugf(ctx, "finding out if the %#q release has to be updated", desiredReleaseState.Name)

	// The release is still being updated so we don't update and check again
	// in the next reconciliation loop.
	if isReleaseInTransitionState(currentReleaseState) {
		upgradeForce := key.HasForceUpgradeAnnotation(cr)
		// Only perform a rollback in case of upgrade force is enabled.
		// This is to consider critical app's service level and stateful apps.
		if upgradeForce {
			err = r.rollback(ctx, obj, currentReleaseState.Status)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			// no-op after rollback
			return nil, nil
		}
	}

	// The release is failed and the values and version have not changed. So we
	// don't update. We will be alerted so we can investigate.
	if isReleaseFailed(currentReleaseState, desiredReleaseState) {
		r.logger.Debugf(ctx, "the %#q release is in status %#q and cannot be updated", desiredReleaseState.Name, currentReleaseState.Status)
		return nil, nil
	}

	// We check the release history and if it has failed more than the max number
	// of attempts we stop updating. Otherwise too many secrets will be created.
	if currentReleaseState.Status == helmclient.StatusFailed {
		isFailedMaxAttempts, err := r.isReleaseFailedMaxAttempts(ctx, key.Namespace(cr), desiredReleaseState.Name)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		if isFailedMaxAttempts {
			r.logger.Debugf(ctx, "the %#q release is in status %#q and has failed %d times", desiredReleaseState.Name, currentReleaseState.Status, releaseFailedMaxAttempts)
			return nil, nil
		}
	}

	if isReleaseModified(currentReleaseState, desiredReleaseState) {
		// Ignoring `Values` in diff since it could contain secret data and we use MD5 hash for comparison.
		opt := cmp.FilterPath(func(p cmp.Path) bool {
			return p.String() == "Values"
		}, cmp.Ignore())

		if diff := cmp.Diff(currentReleaseState, desiredReleaseState, opt); diff != "" {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q has to be updated", cr.Name), "diff", fmt.Sprintf("(-current +desired):\n%s", diff))
		}

		return &desiredReleaseState, nil
	}

	err = r.removeAnnotation(ctx, &cr, annotation.RollbackCount)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return nil, nil
}

func (r *Resource) rollback(ctx context.Context, obj interface{}, currentStatus string) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	count, ok := cr.GetAnnotations()[annotation.RollbackCount]

	var rollbackCount int

	if !ok {
		rollbackCount = 0
	} else {
		rollbackCount, err = strconv.Atoi(count)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if rollbackCount > r.maxRollback {
		r.logger.Debugf(ctx, "the %#q release is in status %#q and has reached max %d rollbacks", key.ReleaseName(cr), currentStatus, r.maxRollback)
		return nil
	}

	if currentStatus == helmclient.StatusPendingInstall {
		r.logger.Debugf(ctx, "deleting release %#q in %#q status", key.ReleaseName(cr), currentStatus)

		err = r.helmClient.DeleteRelease(ctx, key.Namespace(cr), key.ReleaseName(cr))
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted release %#q", key.ReleaseName(cr))
	} else {
		r.logger.Debugf(ctx, "rollback release %#q in %#q status", key.ReleaseName(cr), currentStatus)

		// Rollback to revision 0 restore a release to the previous revision.
		err = r.helmClient.Rollback(ctx, key.Namespace(cr), key.ReleaseName(cr), 0, helmclient.RollbackOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "rollbacked release %#q", key.ReleaseName(cr))
	}

	err = r.addAnnotation(ctx, &cr, annotation.RollbackCount, fmt.Sprintf("%d", rollbackCount+1))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
