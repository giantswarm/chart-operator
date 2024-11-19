package release

import (
	"context"

	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/finalizerskeptcontext"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/resourcecanceledcontext"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/crud"

	"github.com/giantswarm/chart-operator/v4/service/controller/chart/key"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// We use elevated Helm client when performing deletion-wise operations to
	// avoid permissions issues when deleting App Bundles from cluster namespace,
	// see: https://github.com/giantswarm/giantswarm/issues/25731
	hc := r.helmClients.Get(ctx, cr, true)

	releaseState, err := toReleaseState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if releaseState.Name != "" {
		r.logger.Debugf(ctx, "deleting release %#q", releaseState.Name)

		opts := helmclient.DeleteOptions{}
		timeout := key.UninstallTimeout(cr)

		if timeout != nil {
			opts.Timeout = (*timeout).Duration
		}

		err = hc.DeleteRelease(ctx, key.Namespace(cr), releaseState.Name, opts)
		if helmclient.IsReleaseNotFound(err) {
			r.logger.Debugf(ctx, "release %#q already deleted", releaseState.Name)
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		rel, err := hc.GetReleaseContent(ctx, key.Namespace(cr), releaseState.Name)
		if rel != nil {
			// Release still exists. We cancel the resource and keep the finalizer.
			// We will retry the delete in the next reconciliation loop.
			r.logger.Debugf(ctx, "release %#q still exists", releaseState.Name)

			finalizerskeptcontext.SetKept(ctx)
			r.logger.Debugf(ctx, "keeping finalizers")

			resourcecanceledcontext.SetCanceled(ctx)
			r.logger.Debugf(ctx, "canceling resource")

			return nil
		} else if helmclient.IsReleaseNotFound(err) {
			r.logger.Debugf(ctx, "deleted release %#q", releaseState.Name)
		} else if err != nil {
			return microerror.Mask(err)
		}
	} else {
		r.logger.Debugf(ctx, "not deleting release %#q", releaseState.Name)
	}
	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*crud.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := crud.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentReleaseState, err := toReleaseState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredReleaseState, err := toReleaseState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out if the %#q release has to be deleted", desiredReleaseState.Name)

	if !isEmpty(currentReleaseState) && currentReleaseState.Name == desiredReleaseState.Name {
		r.logger.Debugf(ctx, "the %#q release needs to be deleted", desiredReleaseState.Name)

		return &desiredReleaseState, nil
	} else {
		r.logger.Debugf(ctx, "the %#q release does not need to be deleted", desiredReleaseState.Name)
	}

	return nil, nil
}
