package release

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
	yaml "gopkg.in/yaml.v2"
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

	configMapValues, err := r.getConfigMapValues(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretValues, err := r.getSecretValues(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	values, err := union(configMapValues, secretValues)
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

func (r *Resource) getConfigMapValues(ctx context.Context, cr v1alpha1.Chart) (map[string]interface{}, error) {
	configMapValues := make(map[string]interface{})

	if key.IsDeleted(cr) {
		// Return early as configmap has already been deleted.
		return configMapValues, nil
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

		yamlData := configMap.Data[valuesKey]
		if yamlData != "" {
			err = yaml.Unmarshal([]byte(yamlData), &configMapValues)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
	}

	return configMapValues, nil
}

func (r *Resource) getSecretValues(ctx context.Context, cr v1alpha1.Chart) (map[string]interface{}, error) {
	secretValues := make(map[string]interface{})

	if key.IsDeleted(cr) {
		// Return early as secret has already been deleted.
		return secretValues, nil
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

		yamlData := secret.Data[valuesKey]
		if yamlData != nil {
			err = yaml.Unmarshal(yamlData, &secretValues)
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
			// The configmap and secret have at least one shared key. We cannot
			// decide which value should be applied.
			return nil, microerror.Maskf(invalidExecutionError, "configmap and secret share the same key %#q", k)
		}
		a[k] = v
	}
	return a, nil
}
