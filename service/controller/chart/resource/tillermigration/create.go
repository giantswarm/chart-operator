package tillermigration

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/v2/pkg/project"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
)

// EnsureCreated ensures tiller resources are deleted once all helm releases are migrated to v3.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Resource is used to remove tiller. So for other charts we can skip this step.
	if key.ReleaseName(cr) != project.Name() {
		r.logger.Debugf(ctx, "no need to delete tiller for %#q", key.ReleaseName(cr))
		r.logger.Debugf(ctx, "canceling resource")
		return nil
	}

	charts, err := r.g8sClient.ApplicationV1alpha1().Charts("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	notStarted := []string{}
	inProgress := []string{}

	for _, chart := range charts.Items {
		hasConfigMap, err := r.findHelmV2ConfigMaps(ctx, key.ReleaseName(chart))
		if err != nil {
			return microerror.Mask(err)
		}

		hasSecret, err := r.findHelmV3Secrets(ctx, key.ReleaseName(chart), key.Namespace(chart))
		if err != nil {
			return microerror.Mask(err)
		}

		// If Helm v2 release configmap had not been deleted and Helm v3 release secret is there,
		// It means helm release migration is in progress.
		if hasConfigMap && hasSecret {
			inProgress = append(inProgress, chart.Name)
		}

		// If Helm v2 release configmap was not deleted and Helm v3 release secret was not created,
		// It means helm v3 release migration is not started.
		if hasConfigMap && !hasSecret {
			notStarted = append(notStarted, chart.Name)
		}
	}

	if len(notStarted) > 0 || len(inProgress) > 0 {
		// If helm v3 migration was not started or in progress, we could not delete tiller resource.
		r.logger.Debugf(ctx, "following releases are not in migration step; %s", notStarted)
		r.logger.Debugf(ctx, "following releases are in progress migration; %s", inProgress)

		r.logger.Debugf(ctx, "canceling resource.")
		return nil
	}

	r.logger.Debugf(ctx, "no pending or in-progress helm release, deleting all tiller resource")
	err = r.ensureTillerDeleted(ctx)
	if err != nil {
		return microerror.Mask(err)
	}
	r.logger.Debugf(ctx, "deleted all tiller resource")

	return nil
}

func (r *Resource) findHelmV2ConfigMaps(ctx context.Context, releaseName string) (bool, error) {
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "NAME", releaseName, "OWNER", "TILLER"),
	}

	// Check whether it keep helm2 release configMaps
	cms, err := r.k8sClient.CoreV1().ConfigMaps(r.tillerNamespace).List(ctx, lo)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return len(cms.Items) > 0, nil
}

func (r *Resource) findHelmV3Secrets(ctx context.Context, releaseName, releaseNamespace string) (bool, error) {
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "name", releaseName, "owner", "helm"),
	}

	// Check whether it keep helm3 release secrets
	secrets, err := r.k8sClient.CoreV1().Secrets(releaseNamespace).List(ctx, lo)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return len(secrets.Items) > 0, nil
}
