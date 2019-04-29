package release

import (
	"context"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseName := key.ReleaseName(cr)
	releaseContent, err := r.helmClient.GetReleaseContent(ctx, releaseName)
	if helmclient.IsReleaseNotFound(err) {
		// Return early as release is not installed.
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
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
