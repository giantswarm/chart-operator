package chart

import (
	"context"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/service/controller/v5/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	err = r.helmClient.EnsureTillerInstalled(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseName := key.ReleaseName(customObject)
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

	chartState := &ChartState{
		ChannelName:    key.ChannelName(customObject),
		ChartName:      key.ChartName(customObject),
		ChartValues:    releaseContent.Values,
		ReleaseName:    releaseName,
		ReleaseVersion: releaseHistory.Version,
	}

	return chartState, nil
}
