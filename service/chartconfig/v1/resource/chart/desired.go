package chart

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseVersion, err := r.apprClient.GetReleaseVersion(customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ChartState{
		ChartName:      key.ChartName(customObject),
		ChannelName:    key.ChannelName(customObject),
		ReleaseVersion: releaseVersion,
	}

	return chartState, nil
}
