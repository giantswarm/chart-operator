package chartmigration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/chart-operator/pkg/annotation"
	"github.com/giantswarm/chart-operator/pkg/label"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding chartconfig for chart %#q", cr.Name))

	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.App, cr.Labels[label.App]),
	}
	res, err := r.g8sClient.CoreV1alpha1().ChartConfigs(cr.Namespace).List(lo)
	if isChartConfigNotInstalled(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no chartconfig CRD. nothing to do.")
		return nil
	} else if isChartConfigNotAvailable(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", "chartconfig CRs not avaiable.")
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource.")
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	var chartConfig v1alpha1.ChartConfig

	if len(res.Items) == 1 {
		chartConfig = res.Items[0]
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found chartconfig for chart %#q", cr.Name))
	} else if len(res.Items) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no chartconfig CR. nothing to do.")
		return nil
	} else {
		return microerror.Maskf(executionFailedError, "expected 1 chartconfig CR but found %d", len(res.Items))
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking if chartconfig %#q has been migrated", cr.Name))

	if key.HasDeleteCROnlyAnnotation(chartConfig) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig %#q has been migrated", cr.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "removing finalizer")

		err = r.removeFinalizer(ctx, chartConfig)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", "removed finalizer")
	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig %#q has not been migrated", cr.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	return nil
}

// removeFinalizer removes the operatorkit finalizer for the chartconfig CR.
// Finalizers are a JSON array so we need to get the index and remove it using a
// JSON Patch operation.
func (r *Resource) removeFinalizer(ctx context.Context, chartConfig v1alpha1.ChartConfig) error {
	patches := []patch{}

	if len(chartConfig.Finalizers) == 0 {
		// Return early as nothing to do.
		return nil
	}

	var index int

	for i, val := range chartConfig.Finalizers {
		if val == annotation.DeleteCustomResourceOnly {
			index = i
			break
		}
	}

	patches = append(patches, patch{
		Op:   "remove",
		Path: fmt.Sprintf("/metadata/finalizers/%d", index),
	})
	bytes, err := json.Marshal(patches)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.g8sClient.CoreV1alpha1().ChartConfigs(chartConfig.Namespace).Patch(chartConfig.Name, types.JSONPatchType, bytes)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
