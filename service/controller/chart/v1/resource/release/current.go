package release

import (
	"context"
	"fmt"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if key.IsCordoned(cr) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q has been cordoned until %#q due to reason %#q ", key.ReleaseName(cr), key.CordonUntil(cr), key.CordonReason(cr)))

		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil
	}

	releaseName := key.ReleaseName(cr)
	releaseContent, err := r.helmClient.GetReleaseContent(ctx, releaseName)
	if helmclient.IsReleaseNotFound(err) {
		// Return early as release is not installed.
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	if releaseContent.Status == "FAILED" && releaseContent.Name == r.projectName {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not updating release %#q since it's bootstrap from app-operator", releaseContent.Name))

		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil
	}

	releaseHistory, err := r.helmClient.GetReleaseHistory(ctx, releaseName)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseState := &ReleaseState{
		Name:              releaseName,
		Status:            releaseContent.Status,
		ValuesMD5Checksum: key.ValuesMD5ChecksumAnnotation(cr),
		Version:           releaseHistory.Version,
	}

	return releaseState, nil
}
