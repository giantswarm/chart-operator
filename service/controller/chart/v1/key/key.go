package key

import (
	"strconv"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/pkg/annotation"
)

const (
	ChartOperatorName = "chart-operator"

	// ForceHelmUpgradeAnnotationName is the name of the annotation that
	// controls whether force is used when upgrading the Helm release.
	ForceHelmUpgradeAnnotationName = "chart-operator.giantswarm.io/force-helm-upgrade"

	// ValuesMD5ChecksumAnnotationName is the name of the annotation storing
	// an MD5 checksum of the Helm release values.
	ValuesMD5ChecksumAnnotationName = "chart-operator.giantswarm.io/values-md5-checksum"

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

func CordonReason(customResource v1alpha1.Chart) string {
	return customResource.GetAnnotations()[annotation.CordonReason]
}

func CordonUntil(customResource v1alpha1.Chart) string {
	return customResource.GetAnnotations()[annotation.CordonUntilDate]
}

// HasDeleteCROnlyAnnotation returns true if the legacy chartconfig CR has the
// delete custom resource only annotation added by cluster-operator to signal
// that the migration to chart CR is complete.
func HasDeleteCROnlyAnnotation(customResource corev1alpha1.ChartConfig) bool {
	val, ok := customResource.Annotations[annotation.DeleteCustomResourceOnly]
	if !ok {
		return false
	}

	result, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}

	return result
}

func HasForceUpgradeAnnotation(customResource v1alpha1.Chart) bool {
	val, ok := customResource.Annotations[ForceHelmUpgradeAnnotationName]
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
	_, reasonOk := customResource.Annotations[annotation.CordonReason]
	_, untilOk := customResource.Annotations[annotation.CordonUntilDate]

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

// ValuesMD5ChecksumAnnotation returns the annotation value to determine if the
// Helm release values have changed.
func ValuesMD5ChecksumAnnotation(customResource v1alpha1.Chart) string {
	if val, ok := customResource.ObjectMeta.Annotations[ValuesMD5ChecksumAnnotationName]; ok {
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
	if val, ok := customResource.ObjectMeta.Labels[versionLabel]; ok {
		return val
	} else {
		return ""
	}
}
