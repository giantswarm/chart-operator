package chart

import (
	"context"
	"fmt"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"

	"github.com/giantswarm/chart-operator/pkg/annotation"
	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v7/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	reason, reasonOk := customObject.Labels[annotation.CordonReasonAnnotationName]
	until, untilOk := customObject.Labels[annotation.CordonUntilAnnotationName]

	if reasonOk && untilOk {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chart %#q had been cordoned off until %s with following reason; %s ", key.ChartName(customObject), until, reason))

		resourcecanceledcontext.SetCanceled(ctx)
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

		return nil, nil
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
		ReleaseStatus:  releaseContent.Status,
		ReleaseVersion: releaseHistory.Version,
	}

	return chartState, nil
}
