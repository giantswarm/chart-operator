package key

import (
	"strconv"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"

	chartmeta "github.com/giantswarm/chart-operator/v2/pkg/annotation"
)

func AppName(customResource v1alpha1.Chart) string {
	return customResource.GetAnnotations()[annotation.AppName]
}

func AppNamespace(customResource v1alpha1.Chart) string {
	return customResource.GetAnnotations()[annotation.AppNamespace]
}

func ChartStatus(customResource v1alpha1.Chart) v1alpha1.ChartStatus {
	return customResource.Status
}

func SkipCRDs(customResource v1alpha1.Chart) bool {
	return customResource.Spec.Install.SkipCRDs
}

func ConfigMapName(customResource v1alpha1.Chart) string {
	return customResource.Spec.Config.ConfigMap.Name
}

func ConfigMapNamespace(customResource v1alpha1.Chart) string {
	return customResource.Spec.Config.ConfigMap.Namespace
}

func CordonReason(customResource v1alpha1.Chart) string {
	return customResource.GetAnnotations()[chartmeta.CordonReason]
}

func CordonUntil(customResource v1alpha1.Chart) string {
	return customResource.GetAnnotations()[chartmeta.CordonUntilDate]
}

func HasForceUpgradeAnnotation(customResource v1alpha1.Chart) bool {
	val, ok := customResource.Annotations[chartmeta.ForceHelmUpgrade]
	if !ok {
		return false
	}

	result, err := strconv.ParseBool(val)
	if err != nil {
		// If we cannot parse the boolean we return false and this is shown
		// in the logs.
		return false
	}

	return result
}

func IsCordoned(customResource v1alpha1.Chart) bool {
	_, reasonOk := customResource.Annotations[chartmeta.CordonReason]
	_, untilOk := customResource.Annotations[chartmeta.CordonUntilDate]

	if reasonOk && untilOk {
		return true
	} else {
		return false
	}

}

func IsDeleted(customResource v1alpha1.Chart) bool {
	return customResource.GetDeletionTimestamp() != nil
}

func Namespace(customResource v1alpha1.Chart) string {
	return customResource.Spec.Namespace
}

func NamespaceAnnotations(customResource v1alpha1.Chart) map[string]string {
	return customResource.Spec.NamespaceConfig.Annotations
}

func NamespaceLabels(customResource v1alpha1.Chart) map[string]string {
	return customResource.Spec.NamespaceConfig.Labels
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

// ToCustomResource converts value to v1alpha1.Chart and returns it or error
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

// ValuesMD5ChecksumAnnotation returns the annotation value to determine if the
// Helm release values have changed.
func ValuesMD5ChecksumAnnotation(customResource v1alpha1.Chart) string {
	if val, ok := customResource.ObjectMeta.Annotations[chartmeta.ValuesMD5Checksum]; ok {
		return val
	} else {
		return ""
	}
}

func Version(customResource v1alpha1.Chart) string {
	return customResource.Spec.Version
}

// VersionLabel returns the label value to determine if the custom resource is
// supported by this version of the operatorkit resource.
func VersionLabel(customResource v1alpha1.Chart) string {
	if val, ok := customResource.ObjectMeta.Labels[label.ChartOperatorVersion]; ok {
		return val
	} else {
		return ""
	}
}
