package release

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseName := key.ReleaseName(cr)

	tarballURL := key.TarballURL(cr)

	tarballPath, err := r.helmClient.PullChartTarball(ctx, tarballURL)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	defer func() {
		err := r.fs.Remove(tarballPath)
		if err != nil {
			r.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("deletion of %#q failed", tarballPath), "stack", fmt.Sprintf("%#v", err))
		}
	}()

	chart, err := r.helmClient.LoadChart(ctx, tarballPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMapData, err := r.getConfigMapData(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretData, err := r.getSecretData(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	values, err := helmclient.MergeValues(configMapData, secretData)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseState := &ReleaseState{
		Name:    releaseName,
		Status:  helmDeployedStatus,
		Values:  values,
		Version: chart.Version,
	}

	return releaseState, nil
}

func (r *Resource) getConfigMapData(ctx context.Context, cr v1alpha1.Chart) (map[string][]byte, error) {
	configMapData := map[string][]byte{}

	// TODO: Improve desired state generation by removing call to key.IsDeleted.
	//
	//	See https://github.com/giantswarm/giantswarm/issues/5719
	//
	if key.IsDeleted(cr) {
		// Return early as configmap has already been deleted.
		return configMapData, nil
	}

	if key.ConfigMapName(cr) != "" {
		configMapName := key.ConfigMapName(cr)
		configMapNamespace := key.ConfigMapNamespace(cr)

		configMap, err := r.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(configMapName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "config map %#q in namespace %#q not found", configMapName, configMapNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		for k, v := range configMap.Data {
			configMapData[k] = []byte(v)
		}
	}

	return configMapData, nil
}

func (r *Resource) getSecretData(ctx context.Context, cr v1alpha1.Chart) (map[string][]byte, error) {
	secretData := map[string][]byte{}

	// TODO: Improve desired state generation by removing call to key.IsDeleted.
	//
	//	See https://github.com/giantswarm/giantswarm/issues/5719
	//
	if key.IsDeleted(cr) {
		// Return early as secret has already been deleted.
		return secretData, nil
	}

	if key.SecretName(cr) != "" {
		secretName := key.SecretName(cr)
		secretNamespace := key.SecretNamespace(cr)

		secret, err := r.k8sClient.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "secret %#q in namespace %#q not found", secretName, secretNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		secretData = secret.Data
	}

	return secretData, nil
}
