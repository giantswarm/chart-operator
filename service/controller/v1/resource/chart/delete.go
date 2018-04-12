package chart

import (
	"context"
	"fmt"
	"reflect"

	"github.com/giantswarm/chart-operator/service/controller/v1/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	chartState, err := toChartState(deleteChange)
	if err != nil {
		return microerror.Mask(err)
	}

	if chartState.ReleaseName != "" {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting release %s", chartState.ReleaseName))
		release := key.ReleaseName(customObject)

		err := r.helmClient.DeleteRelease(release)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting release %s", chartState.ReleaseName))
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("not deleting release %s", chartState.ReleaseName))
	}
	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	delete, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
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

	if !reflect.DeepEqual(currentChartState, ChartState{}) && reflect.DeepEqual(currentChartState, desiredChartState) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %s release needs to be deleted", desiredChartState.ReleaseName))

		return &desiredChartState, nil
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("the %s release does not need to be deleted", desiredChartState.ReleaseName))
	}

	return nil, nil
}
