package releasemigration

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/reconciliationcanceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

// EnsureCreated ensures helm release is migrated from a v2 configmap to a v3 secret.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	hasConfigMap, err := r.hasHelmV2ConfigMaps(ctx, key.ReleaseName(cr))
	if err != nil {
		return microerror.Mask(err)
	}

	hasSecret, err := r.hasHelmV3Secrets(ctx, key.ReleaseName(cr), key.Namespace(cr))
	if err != nil {
		return microerror.Mask(err)
	}

	// If Helm v2 release configmap had not been deleted and Helm v3 release secret is there,
	// It means helm release migration is in progress.
	if hasConfigMap && hasSecret {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q helmV3 migration in progress", key.ReleaseName(cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
		return nil
	}

	// If Helm v2 release configmap had not been deleted and Helm v3 release secret was not created,
	// It means helm v3 release migration is not started.
	if hasConfigMap && !hasSecret {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q helmV3 migration not started", key.ReleaseName(cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")

		// install helm-2to3-migration app
		err := r.ensureReleasesMigrated(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("no pending migration for release %#q", key.ReleaseName(cr)))

	return nil
}

func (r *Resource) hasHelmV2ConfigMaps(ctx context.Context, releaseName string) (bool, error) {
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "NAME", releaseName, "OWNER", "TILLER"),
	}

	// Check whether helm 2 release configMaps still exist.
	cms, err := r.k8sClient.CoreV1().ConfigMaps(r.tillerNamespace).List(lo)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return len(cms.Items) > 0, nil
}

func (r *Resource) hasHelmV3Secrets(ctx context.Context, releaseName, releaseNamespace string) (bool, error) {
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "name", releaseName, "owner", "helm"),
	}

	// Check whether helm 3 release secret exists.
	secrets, err := r.k8sClient.CoreV1().Secrets(releaseNamespace).List(lo)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return len(secrets.Items) > 0, nil
}
