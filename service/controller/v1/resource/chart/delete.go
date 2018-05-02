package chart

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	chartState, err := toChartState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if chartState.ReleaseName != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting release %s", chartState.ReleaseName))

		err := r.helmClient.DeleteRelease(chartState.ReleaseName)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting release %s", chartState.ReleaseName))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting release %s", chartState.ReleaseName))
	}
	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := controller.NewPatch()
	patch.SetDeleteChange(delete)

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentChartState, err := toChartState(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	desiredChartState, err := toChartState(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding out if the %s release has to be deleted", desiredChartState.ReleaseName))

	if !reflect.DeepEqual(currentChartState, ChartState{}) && currentChartState.ReleaseName == desiredChartState.ReleaseName && currentChartState.ReleaseVersion == desiredChartState.ReleaseVersion {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %s release needs to be deleted", desiredChartState.ReleaseName))

		return &desiredChartState, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %s release does not need to be deleted", desiredChartState.ReleaseName))
	}

	return nil, nil
}
