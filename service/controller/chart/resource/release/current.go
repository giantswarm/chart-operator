package release

import (
	"context"
	"fmt"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/chart-operator/pkg/project"
	"github.com/giantswarm/chart-operator/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/service/controller/chart/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if key.IsCordoned(cr) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q has been cordoned until %#q due to reason %#q ", key.ReleaseName(cr), key.CordonUntil(cr), key.CordonReason(cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	hasConfigmap, err := r.findHelmV2ConfigMaps(ctx, key.ReleaseName(cr))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if hasConfigmap {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q has been not migrated from helm 2", key.ReleaseName(cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	releaseName := key.ReleaseName(cr)
	releaseContent, err := r.helmClient.GetReleaseContent(ctx, key.Namespace(cr), releaseName)
	if helmclient.IsReleaseNotFound(err) {
		// Return early as release is not installed.
		return nil, nil
	} else if helmclient.IsReleaseNameInvalid(err) {
		reason := fmt.Sprintf("release name %#q is invalid", releaseName)
		addStatusToContext(cc, reason, releaseNotInstalledStatus)

		r.logger.LogCtx(ctx, "level", "warning", "message", reason, "stack", microerror.JSON(err))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	if releaseContent.Status == helmclient.StatusFailed && releaseContent.Name == project.Name() {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not updating own release %#q since it's %#q", releaseContent.Name, releaseContent.Status))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	releaseState := &ReleaseState{
		Name:              releaseName,
		Status:            releaseContent.Status,
		ValuesMD5Checksum: key.ValuesMD5ChecksumAnnotation(cr),
		Version:           releaseContent.Version,
	}

	return releaseState, nil
}
