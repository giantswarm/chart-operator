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

	hasConfigMap, err := r.getHelmV2ConfigMaps(ctx, key.ReleaseName(cr))
	if err != nil {
		return microerror.Mask(err)
	}

	hasSecret, err := r.getHelmV3Secrets(ctx, key.ReleaseName(cr), key.Namespace(cr))
	if err != nil {
		return microerror.Mask(err)
	}

	// If Helm v2 release configmap had not been deleted and Helm v3 release secret is there,
	// It means helm release migration is in progress.
	if hasConfigMap && hasSecret {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q is under helmV3 migration", key.ReleaseName(cr)))
		r.logger.LogCtx(ctx, "level", "debug", "message", "canceling reconciliation")
		reconciliationcanceledcontext.SetCanceled(ctx)
		return nil
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("no pending migration on release %#q", key.ReleaseName(cr)))

	return nil
}

func (r *Resource) getHelmV2ConfigMaps(ctx context.Context, releaseName string) (bool, error) {
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "NAME", releaseName, "OWNER", "TILLER"),
	}

	// Check whether it keep helm2 release configMaps
	cms, err := r.k8sClient.CoreV1().ConfigMaps(r.tillerNamespace).List(lo)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return len(cms.Items) > 0, nil
}

func (r *Resource) getHelmV3Secrets(ctx context.Context, releaseName, releaseNamespace string) (bool, error) {
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "name", releaseName, "owner", "helm"),
	}

	// Check whether it keep helm3 release secrets
	secrets, err := r.k8sClient.CoreV1().Secrets(releaseNamespace).List(lo)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return len(secrets.Items) > 0, nil
}
