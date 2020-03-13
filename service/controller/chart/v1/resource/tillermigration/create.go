package tillermigration

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

// EnsureCreated ensures tiller resources are deleted after all helm releases migrated into v3.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Resource is used to remove tiller pod in tenant clusters.
	// So for other charts we can skip this step.
	if key.ReleaseName(cr) != key.ChartOperatorName {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("no need to delete a tiller for %#q", key.ReleaseName(cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
		return nil
	}

	charts, err := r.g8sClient.ApplicationV1alpha1().Charts("giantswarm").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	notStarted := []string{}
	inProgress := []string{}
	for _, chart := range charts.Items {
		foundConfigMap, err := r.findHelmV2ConfigMaps(ctx, chart)
		if err != nil {
			return microerror.Mask(err)
		}

		foundSecret, err := r.findHelmV3Secrets(ctx, chart)
		if err != nil {
			return microerror.Mask(err)
		}

		if foundConfigMap && foundSecret {
			inProgress = append(inProgress, chart.Name)
		} else if foundConfigMap && !foundSecret {
			notStarted = append(notStarted, chart.Name)
		}
	}

	if len(notStarted) > 0 || len(inProgress) > 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Following releases are not in migration step; %s", notStarted))
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("Following releases are in progress migration; %s", inProgress))

		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource.")
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", "no pending or in-progress helm release, deleting all tiller resource")
	err = r.ensureTillerDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.LogCtx(ctx, "level", "debug", "message", "deleted all tiller resource")

	return nil
}

func (r *Resource) findHelmV2ConfigMaps(ctx context.Context, chart v1alpha1.Chart) (bool, error) {
	releaseName := key.ReleaseName(chart)
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "NAME", releaseName, "OWNER", "TILLER"),
	}

	// Check whether it keep helm2 release configMaps
	cms, err := r.k8sClient.CoreV1().ConfigMaps("giantswarm").List(lo)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return len(cms.Items) > 0, nil
}

func (r *Resource) findHelmV3Secrets(ctx context.Context, chart v1alpha1.Chart) (bool, error) {
	releaseName := key.ReleaseName(chart)
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "name", releaseName, "owner", "helm"),
	}

	// Check whether it keep helm3 release secrets
	secrets, err := r.k8sClient.CoreV1().Secrets(key.Namespace(chart)).List(lo)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return len(secrets.Items) > 0, nil
}
