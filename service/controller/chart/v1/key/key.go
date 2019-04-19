package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/microerror"
)

const (
	versionLabel = "chart-operator.giantswarm.io/version"
)

func ChartStatus(customResource v1alpha1.Chart) v1alpha1.ChartStatus {
	return customResource.Status
}

func ConfigMapName(customResource v1alpha1.Chart) string {
	return customResource.Spec.Config.ConfigMap.Name
}

func ConfigMapNamespace(customResource v1alpha1.Chart) string {
	return customResource.Spec.Config.ConfigMap.Namespace
}

func IsDeleted(customResource v1alpha1.Chart) bool {
	return customResource.GetDeletionTimestamp() != nil
}

func Namespace(customResource v1alpha1.Chart) string {
	return customResource.Spec.Namespace
}

func ReleaseName(customResource v1alpha1.Chart) string {
	return customResource.Spec.Name
}

func SecretName(customResource v1alpha1.Chart) string {
	return customResource.Spec.Config.Secret.Name
}

func SecretNamespace(customResource v1alpha1.Chart) string {
	return customResource.Spec.Config.Secret.Namespace
}

func TarballURL(customResource v1alpha1.Chart) string {
	return customResource.Spec.TarballURL
}

// ToCustomResource converts value to v1alpha1.ChartConfig and returns it or error
// if type does not match.
func ToCustomResource(v interface{}) (v1alpha1.Chart, error) {
	customResourcePointer, ok := v.(*v1alpha1.Chart)
	if !ok {
		return v1alpha1.Chart{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.Chart{}, v)
	}

	if customResourcePointer == nil {
		return v1alpha1.Chart{}, microerror.Maskf(emptyValueError, "empty value cannot be converted to customResource")
	}

	return *customResourcePointer, nil
}

func ValuesMD5Checksum(customResource v1alpha1.Chart) string {
	return customResource.Status.Values.MD5Checksum
}

// VersionLabel returns the label value to determine if the custom resource is
// supported by this version of the operatorkit resource.
func VersionLabel(customResource v1alpha1.Chart) string {
	if val, ok := customResource.ObjectMeta.Labels[versionLabel]; ok {
		return val
	} else {
		return ""
	}
}
