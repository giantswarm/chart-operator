package chartmigration

import (
	"context"
	"fmt"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding chartconfig %#q", cr.Name))

	chartConfig, err := r.g8sClient.CoreV1alpha1().ChartConfigs(cr.Namespace).Get(cr.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not find chartconfig %#q. nothing to do.", cr.Name))
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found chartconfig %#q", cr.Name))

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking if chartconfig %#q has been migrated", cr.Name))

	if key.HasDeleteCROnlyAnnotation(chartConfig) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("chartconfig %#q has been migrated", cr.Name))
		r.logger.LogCtx(ctx, "level", "debug", "message", "removing finalizer")

		finalizers := []string{}

		for _, f := range chartConfig.ObjectMeta.Finalizers {
			if f == "operatorkit.giantswarm.io/chart-operator-chartconfig" {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("removing finalizer %#q", f))
			} else {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("keeping finalizer %#q", f))
				finalizers = append(finalizers, f)
			}
		}

		chartConfig.ObjectMeta.Finalizers = finalizers
		_, err = r.g8sClient.CoreV1alpha1().ChartConfigs("giantswarm").Update(chartConfig)
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
