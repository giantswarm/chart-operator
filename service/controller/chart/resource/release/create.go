package release

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/chart-operator/v4/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/v4/service/controller/chart/key"
)

const (
	// subjectToTwoStepInstall marks app (Helm Chart) as needing two step
	// installation process.
	subjectToTwoStepInstall = "application.giantswarm.io/two-step-install"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	hc := r.helmClients.Get(ctx, cr, false)

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	releaseState, err := toReleaseState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if releaseState.Name == "" {
		// no-op
		return nil
	}

	r.logger.Debugf(ctx, "creating release %#q in namespace %#q", releaseState.Name, key.Namespace(cr))

	ns := key.Namespace(cr)
	tarballURL := key.TarballURL(cr)
	skipCRDs := key.SkipCRDs(cr)
	timeout := key.InstallTimeout(cr)

	tarballPath, err := hc.PullChartTarball(ctx, tarballURL)
	if helmclient.IsPullChartFailedError(err) {
		reason := fmt.Sprintf("pulling chart %#q failed", tarballURL)
		addStatusToContext(cc, reason, chartPullFailedStatus)

		r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if helmclient.IsPullChartNotFound(err) {
		reason := fmt.Sprintf("chart %#q not found", tarballURL)
		addStatusToContext(cc, reason, chartPullFailedStatus)

		r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	} else if helmclient.IsPullChartTimeout(err) {
		reason := fmt.Sprintf("timeout pulling %#q", tarballURL)
		addStatusToContext(cc, reason, chartPullFailedStatus)

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

	ch := make(chan error)

	// We create the helm release but with a wait timeout so we don't
	// block reconciling other CRs.
	//
	// If we do timeout the install will continue in the background.
	// We will check the progress in the next reconciliation loop.
	go func() {
		var e error

		defer close(ch)

		if skipCRDs {
			r.logger.Debugf(ctx, "helm release %#q has SkipCRDs set to true, not installing CRDs", releaseState.Name)
		}
		iOpts := helmclient.InstallOptions{
			ReleaseName: releaseState.Name,
			SkipCRDs:    skipCRDs,
		}

		// If the timeout is provided use it
		if timeout != nil {
			r.logger.Debugf(ctx, "using custom %#q timeout to install release %#q", (*timeout).Duration, releaseState.Name)
			iOpts.Timeout = (*timeout).Duration
		}

		// We need to pass the ValueOverrides option to make the install process
		// use the default values and prevent errors on nested values.
		e = hc.InstallReleaseFromTarball(ctx, tarballPath, ns, releaseState.Values, iOpts)

		// We check the error here to return early if installation failed. There is no point
		// in upgrading in such scenario.
		if e != nil {
			err = e
			return
		}

		// Load the chart to get its annotations and verify it is a subject to
		// internal upgrade procedure. If we experience error here we log it and
		// return.
		chart, e := hc.LoadChart(ctx, tarballPath)
		if e != nil {
			r.logger.Errorf(ctx, err, "loading chart %#q failed on internal upgrade", tarballPath)
			return
		}
		if _, ok := chart.Annotations[subjectToTwoStepInstall]; !ok {
			return
		}

		// Hooks get disabled on internal upgrade to make it faster.
		uOpts := helmclient.UpdateOptions{
			DisableHooks: true,
			Force:        false,
		}

		// Internal upgrade gets the same timeout option as installation, as logically
		// it is part of the installation procedure.
		if timeout != nil {
			r.logger.Debugf(ctx, "using custom %#q timeout to internally update release %#q", (*timeout).Duration, releaseState.Name)
			uOpts.Timeout = (*timeout).Duration
		}

		r.logger.Debugf(ctx, "doing internal upgrade for release %#q", releaseState.Name)

		err = hc.UpdateReleaseFromTarball(ctx,
			tarballPath,
			ns,
			releaseState.Name,
			releaseState.Values,
			uOpts)
	}()

	select {
	case <-ch:
		// Fall through.
	case <-time.After(r.k8sWaitTimeout):
		r.logger.Debugf(ctx, "waited for %d secs. release still being created", int64(r.k8sWaitTimeout.Seconds()))

		// We set the hash annotation so the update state calculation is accurate
		// when we check in the next reconciliation loop.
		err = r.addHashAnnotation(ctx, cr, releaseState)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}
	if helmclient.IsResourceAlreadyExists(err) {
		reason := err.Error()
		reason = fmt.Sprintf("object already exists: (%s)", reason)
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

		if isSchemaValidationError(err) {
			r.logger.Errorf(ctx, err, "values schema validation for %#q failed", releaseState.Name)
			addStatusToContext(cc, err.Error(), valuesSchemaViolation)

			r.logger.Debugf(ctx, "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		}

		releaseContent, relErr := hc.GetReleaseContent(ctx, ns, releaseState.Name)
		if helmclient.IsReleaseNotFound(relErr) {
			addStatusToContext(cc, err.Error(), releaseNotInstalledStatus)

			r.logger.Debugf(ctx, "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		} else if relErr != nil {
			return microerror.Mask(relErr)
		}

		// Release is failed so the status resource will check the Helm release.
		if releaseContent.Status == helmclient.StatusFailed {
			addStatusToContext(cc, releaseContent.Description, helmclient.StatusFailed)

			r.logger.Debugf(ctx, "failed to create release %#q", releaseContent.Name)
			r.logger.Debugf(ctx, "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil
		}

		addStatusToContext(cc, err.Error(), unknownError)

		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil
	}

	r.logger.Debugf(ctx, "created release %#q in namespace %#q", releaseState.Name, key.Namespace(cr))

	// We set the hash annotation so the update state calculation
	// is accurate when we check in the next reconciliation loop.
	err = r.addHashAnnotation(ctx, cr, releaseState)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentReleaseState, err := toReleaseState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredReleaseState, err := toReleaseState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out if the %#q release has to be created", desiredReleaseState.Name)

	createState := &ReleaseState{}

	if isEmpty(currentReleaseState) {
		r.logger.Debugf(ctx, "the %#q release needs to be created", desiredReleaseState.Name)

		createState = &desiredReleaseState
	} else {
		r.logger.Debugf(ctx, "the %#q release does not need to be created", desiredReleaseState.Name)
	}

	return createState, nil
}
