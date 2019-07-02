package key

import (
	"strconv"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/pkg/annotation"
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

func CordonReason(customObject v1alpha1.ChartConfig) string {
	val, ok := customObject.Annotations[annotation.CordonReasonAnnotationName]
	if ok {
		return val
	}
	return ""
}

func CordonUntil(customObject v1alpha1.ChartConfig) string {
	val, ok := customObject.Annotations[annotation.CordonUntilAnnotationName]
	if ok {
		return val
	}
	return ""
}

func HasForceUpgradeAnnotation(customObject v1alpha1.ChartConfig) (bool, error) {
	val, ok := customObject.Annotations["chart-operator.giantswarm.io/force-helm-upgrade"]
	if !ok {
		return false, nil
	}

	result, err := strconv.ParseBool(val)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return result, nil
}

func IsCordoned(customObject v1alpha1.ChartConfig) bool {
	_, reasonOk := customObject.Annotations[annotation.CordonReasonAnnotationName]
	_, untilOk := customObject.Annotations[annotation.CordonUntilAnnotationName]

	if reasonOk && untilOk {
		return true
	} else {
		return false
	}

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

func ReleaseStatus(customObject v1alpha1.ChartConfig) string {
	return customObject.Status.ReleaseStatus
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

func UserConfigMapName(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.Chart.UserConfigMap.Name
}

func UserConfigMapNamespace(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.Chart.UserConfigMap.Namespace
}

func VersionBundleVersion(customObject v1alpha1.ChartConfig) string {
	return customObject.Spec.VersionBundle.Version
}
