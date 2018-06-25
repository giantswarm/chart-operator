package chart

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/v2/key"
)

const (
	ValuesData = "values.json"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	customObject, err := key.ToCustomObject(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	name := key.ChartName(customObject)
	channel := key.ChannelName(customObject)
	chartValues, err := r.getConfigMapValues(ctx, customObject)
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
	chartValues := map[string]interface{}{}

	if key.ConfigMapName(customObject) != "" {
		configMapName := key.ConfigMapName(customObject)
		configMapNamespace := key.ConfigMapNamespace(customObject)

		configMap, err := r.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(configMapName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return chartValues, microerror.Maskf(notFoundError, "config map '%s' not found", configMapName)
		} else if err != nil {
			return chartValues, microerror.Mask(err)
		}

		values := configMap.Data[ValuesData]
		r.logger.LogCtx(ctx, "found data %q for config map %q", values, configMapName)
	}

	return chartValues, nil
}
