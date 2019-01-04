package release

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseName := key.ReleaseName(customObject)
	tarballURL := key.TarballURL(customObject)

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

	configMapValues, err := r.getConfigMapValues(ctx, customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretValues, err := r.getSecretValues(ctx, customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	values, err := union(configMapValues, secretValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	releaseState := &ReleaseState{
		Name:    releaseName,
		Status:  "DEPLOYED",
		Values:  values,
		Version: chart.Version,
	}

	return releaseState, nil
}

func (r *Resource) getConfigMapValues(ctx context.Context, customObject v1alpha1.Chart) (map[string]interface{}, error) {
	configMapValues := make(map[string]interface{})

	if key.ConfigMapName(customObject) != "" {
		configMapName := key.ConfigMapName(customObject)
		configMapNamespace := key.ConfigMapNamespace(customObject)

		configMap, err := r.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(configMapName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "config map %#q in namespace %#q not found", configMapName, configMapNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		configMapData := configMap.Data[controller.ConfigMapValuesKey]
		err = json.Unmarshal([]byte(configMapData), &configMapValues)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return configMapValues, nil
}

func (r *Resource) getSecretValues(ctx context.Context, customObject v1alpha1.Chart) (map[string]interface{}, error) {
	secretValues := make(map[string]interface{})

	if key.SecretName(customObject) != "" {
		secretName := key.SecretName(customObject)
		secretNamespace := key.SecretNamespace(customObject)

		secret, err := r.k8sClient.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "secret %#q in namespace %#q not found", secretName, secretNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		secretData := secret.Data[controller.SecretValuesKey]
		if secretData != nil {
			err = json.Unmarshal(secretData, &secretValues)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
	}

	return secretValues, nil
}

func union(a, b map[string]interface{}) (map[string]interface{}, error) {
	if a == nil {
		return b, nil
	}

	for k, v := range b {
		_, ok := a[k]
		if ok {
			// The secret and config map we use have at least one shared key. We can not
			// decide which value is supposed to be applied.
			return nil, microerror.Maskf(invalidConfigError, "secret and config map share the same key %s", k)
		}
		a[k] = v
	}
	return a, nil
}
