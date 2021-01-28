package release

import (
	"context"
	"fmt"

	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v4/pkg/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/chart-operator/v2/pkg/project"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
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
		r.logger.Debugf(ctx, "release %#q has been cordoned until %#q due to reason %#q ", key.ReleaseName(cr), key.CordonUntil(cr), key.CordonReason(cr))
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	hasConfigmap, err := r.findHelmV2ConfigMaps(ctx, key.ReleaseName(cr))
	if err != nil {
		reason := fmt.Sprintf("release %#q didn't migrate to helm 3", key.ReleaseName(cr))
		addStatusToContext(cc, reason, releaseNotInstalledStatus)
		return nil, microerror.Mask(err)
	}

	if hasConfigmap {
		r.logger.Debugf(ctx, "release %#q has not been migrated from helm 2", key.ReleaseName(cr))
		r.logger.Debugf(ctx, "canceling resource")
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
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil

	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	if releaseContent.Status == helmclient.StatusFailed && releaseContent.Name == project.Name() {
		r.logger.Debugf(ctx, "not updating own release %#q since it's %#q", releaseContent.Name, releaseContent.Status)
		r.logger.Debugf(ctx, "canceling resource")
		resourcecanceledcontext.SetCanceled(ctx)
		return nil, nil
	}

	releaseState := &ReleaseState{
		Name:              releaseName,
		Status:            releaseContent.Status,
		ValuesMD5Checksum: key.ValuesMD5ChecksumAnnotation(cr),
		Version:           releaseContent.Version,
	}

	releaseHistory, err := r.helmClient.GetReleaseHistory(ctx, key.Namespace(cr), releaseName)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for i, history := range releaseHistory {
		r.logger.Debugf(ctx, "history: %d %#v", i, history)
	}

	return releaseState, nil
}
