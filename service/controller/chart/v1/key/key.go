package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	versionBundleAnnotation = "giantswarm.io/version-bundle"
)

func ConfigMapName(customObject v1alpha1.Chart) string {
	return customObject.Spec.Config.ConfigMap.Name
}

func ConfigMapNamespace(customObject v1alpha1.Chart) string {
	return customObject.Spec.Config.ConfigMap.Namespace
}

func Namespace(customObject v1alpha1.Chart) string {
	return customObject.Spec.Namespace
}

func ReleaseName(customObject v1alpha1.Chart) string {
	return customObject.Spec.Name
}

func SecretName(customObject v1alpha1.Chart) string {
	return customObject.Spec.Config.Secret.Name
}

func SecretNamespace(customObject v1alpha1.Chart) string {
	return customObject.Spec.Config.Secret.Namespace
}

func TarballURL(customObject v1alpha1.Chart) string {
	return customObject.Spec.TarballURL
}

// ToCustomObject converts value to v1alpha1.ChartConfig and returns it or error
// if type does not match.
func ToCustomObject(v interface{}) (v1alpha1.Chart, error) {
	customObjectPointer, ok := v.(*v1alpha1.Chart)
	if !ok {
		return v1alpha1.Chart{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.Chart{}, v)
	}

	if customObjectPointer == nil {
		return v1alpha1.Chart{}, microerror.Maskf(emptyValueError, "empty value cannot be converted to CustomObject")
	}

	return *customObjectPointer, nil
}

func VersionBundleVersion(customObject v1alpha1.Chart) string {
	if val, ok := customObject.ObjectMeta.Annotations[versionBundleAnnotation]; ok {
		return val
	} else {
		return ""
	}
}
