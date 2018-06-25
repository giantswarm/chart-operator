package chart

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/service/controller/v2/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	name := key.ChartName(customObject)
	channel := key.ChannelName(customObject)
	releaseVersion, err := r.apprClient.GetReleaseVersion(name, channel)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ChartState{
		ChartName:      key.ChartName(customObject),
		ChannelName:    key.ChannelName(customObject),
		ReleaseName:    key.ReleaseName(customObject),
		ReleaseVersion: releaseVersion,
	}

	return chartState, nil
}
