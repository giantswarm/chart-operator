package chart

import (
	"context"
	"encoding/json"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/v3/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	name := key.ChartName(customObject)
	channel := key.ChannelName(customObject)
	chartConfigmapValues, err := r.getConfigMapValues(ctx, customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	chartSecretValues, err := r.getSecretValues(ctx, customObject)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	chartValues, err := union(chartConfigmapValues, chartSecretValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	releaseVersion, err := r.apprClient.GetReleaseVersion(name, channel)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	chartState := &ChartState{
		ChartName:      key.ChartName(customObject),
		ChartValues:    chartValues,
		ChannelName:    key.ChannelName(customObject),
		ReleaseName:    key.ReleaseName(customObject),
		ReleaseVersion: releaseVersion,
	}

	return chartState, nil
}

func (r *Resource) getConfigMapValues(ctx context.Context, customObject v1alpha1.ChartConfig) (map[string]interface{}, error) {
	chartValues := make(map[string]interface{})

	if key.ConfigMapName(customObject) != "" {
		configMapName := key.ConfigMapName(customObject)
		configMapNamespace := key.ConfigMapNamespace(customObject)

		configMap, err := r.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(configMapName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return chartValues, microerror.Maskf(notFoundError, "config map '%s' in namespace '%s' not found", configMapName, configMapNamespace)
		} else if err != nil {
			return chartValues, microerror.Mask(err)
		}

		jsonData := configMap.Data["values.json"]
		if jsonData != "" {
			err = json.Unmarshal([]byte(jsonData), &chartValues)
			if err != nil {
				return chartValues, microerror.Mask(err)
			}
		}
	}

	return chartValues, nil
}

func (r *Resource) getSecretValues(ctx context.Context, customObject v1alpha1.ChartConfig) (map[string]interface{}, error) {
	secretValues := make(map[string]interface{})

	if key.SecretName(customObject) != "" {
		secretName := key.SecretName(customObject)
		secretNamespace := key.SecretNamespace(customObject)

		secret, err := r.k8sClient.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "secret '%s' in namespace '%s' not found", secretName, secretNamespace)
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		// TODO: fix this "secret.json" name somewhere and access it in release-operator.
		secretData := secret.Data["secret.json"]
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
