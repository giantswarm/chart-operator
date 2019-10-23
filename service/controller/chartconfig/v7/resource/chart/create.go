package chart

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v7/key"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	chartState, err := toChartState(createChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if chartState.ChartName != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating chart %#q", chartState.ChartName))
		name := key.ChartName(customObject)
		channel := key.ChannelName(customObject)
		ns := key.Namespace(customObject)

		tarballPath, err := r.apprClient.PullChartTarball(ctx, name, channel)
		if err != nil {
			return microerror.Mask(err)
		}
		defer func() {
			err := r.fs.Remove(tarballPath)
			if err != nil {
				r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %q failed", tarballPath), "stack", fmt.Sprintf("%#v", err))
			}
		}()

		values, err := json.Marshal(chartState.ChartValues)
		if err != nil {
			return microerror.Mask(err)
		}

		// We need to pass the ValueOverrides option to make the install process
		// use the default values and prevent errors on nested values.
		//
		//     {
		//      rpc error: code = Unknown desc = render error in "cnr-server-chart/templates/deployment.yaml":
		//      template: cnr-server-chart/templates/deployment.yaml:20:26:
		//      executing "cnr-server-chart/templates/deployment.yaml" at <.Values.image.reposi...>: can't evaluate field repository in type interface {}
		//     }
		//
		err = r.helmClient.InstallReleaseFromTarball(ctx, tarballPath, ns, helm.ReleaseName(chartState.ReleaseName), helm.ValueOverrides(values))
		if err != nil {
			releaseContent, err := r.helmClient.GetReleaseContent(ctx, chartState.ReleaseName)
			if err != nil {
				return microerror.Mask(err)
			}
			if releaseContent.Status == releaseStatusFailed {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("failed to update release %#q", releaseContent.Name))

				resourcecanceledcontext.SetCanceled(ctx)
				r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")

				return nil
			}
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created chart %#q", chartState.ChartName))
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentChartState, err := toChartState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartState, err := toChartState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if the %#q chart has to be created", desiredChartState.ChartName))

	createState := &ChartState{}

	if currentChartState.IsEmpty() {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q chart needs to be created", desiredChartState.ChartName))

		createState = &desiredChartState
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %#q chart does not need to be created", desiredChartState.ChartName))
	}

	return createState, nil
}
