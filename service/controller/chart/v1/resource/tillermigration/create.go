package tillermigration

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

// EnsureCreated ensures Tiller is installed and the latest version.
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

	notDeleted := map[string]bool{}
	inProgress := map[string]bool{}
	for _, chart := range charts.Items {
		releaseName := key.ReleaseName(chart)
		releaseNamespace := key.Namespace(chart)
		lo := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "NAME", releaseName, "OWNER", "TILLER"),
		}

		// Check whether it keep helm2 release configMaps
		cms, err := r.k8sClient.CoreV1().ConfigMaps(releaseNamespace).List(lo)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(cms.Items) > 0 {
			notDeleted[releaseName] = true
		}

		lo = metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "name", releaseName, "owner", "helm"),
		}

		// Check whether it keep helm3 release secrets
		secrets, err := r.k8sClient.CoreV1().Secrets(releaseNamespace).List(lo)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(secrets.Items) > 0 {
			inProgress[releaseName] = true
		}
	}

	if len(notDeleted) == 0 {
		r.logger.LogCtx(ctx, "level", "debug", "message", "no pending or in-progress helm release, deleting all tiller resource")
		err := r.ensureTillerDeleted(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		r.logger.LogCtx(ctx, "level", "debug", "message", "deleted all tiller resource")
		return nil
	}

	for name, _ := range notDeleted {
		if inProgress[name] {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("migration of release %#q is still in progress", name))
		} else {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("migration of release %#q is not started", name))
		}
	}

	return nil
}
