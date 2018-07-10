package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

func ChartName(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.Chart.Name
}

func ChannelName(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.Chart.Channel
}

func ConfigMapName(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.Chart.ConfigMap.Name
}

func ConfigMapNamespace(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.Chart.ConfigMap.Namespace
}

func Namespace(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.Chart.Namespace
}

func SecretName(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.Chart.Secret.Name
}

func SecretNamespace(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.Chart.Secret.Namespace
}

func ReleaseName(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.Chart.Release
}

// ToCustomObject converts value to v1alpha1.ChartConfig and returns it or error
// if type does not match.
func ToCustomObject(v interface{}) (v1alpha1.ChartConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.ChartConfig)
	if !ok {
		return v1alpha1.ChartConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.ChartConfig{}, v)
	}

	if customObjectPointer == nil {
		return v1alpha1.ChartConfig{}, microerror.Maskf(emptyValueError, "empty value cannot be converted to CustomObject")
	}

	return *customObjectPointer, nil
}

func VersionBundleVersion(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.VersionBundle.Version
}
