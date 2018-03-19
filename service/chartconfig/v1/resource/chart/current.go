package chart

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/helm"
	"github.com/giantswarm/chart-operator/service/chartconfig/v1/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseContent, err := r.helmClient.GetReleaseContent(customObject)
	if helm.IsReleaseNotFound(err) {
		// Fall through.
		return &ChartState{}, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseHistory, err := r.helmClient.GetReleaseHistory(customObject)
	if helm.IsReleaseNotFound(err) {
		// Fall through.
		return &ChartState{}, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ChartState{
		ChartName:      key.ChartName(customObject),
		ChannelName:    key.ChannelName(customObject),
		ReleaseName:    key.ReleaseName(customObject),
		ChartValues:    releaseContent.Values,
		ReleaseStatus:  releaseContent.Status,
		ReleaseVersion: releaseHistory.Version,
	}

	return chartState, nil
}
