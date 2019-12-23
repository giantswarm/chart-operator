package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"

	"github.com/giantswarm/chart-operator/pkg/annotation"
)

func ChartName(customResource v1alpha1.ChartConfig) string {
	return customResource.Spec.Chart.Name
}

func ChannelName(customResource v1alpha1.ChartConfig) string {
	return customResource.Spec.Chart.Channel
}

func CordonReason(customResource v1alpha1.ChartConfig) string {
	return customResource.GetAnnotations()[annotation.CordonReason]
}

func CordonUntil(customResource v1alpha1.ChartConfig) string {
	return customResource.GetAnnotations()[annotation.CordonUntilDate]
}

func IsCordoned(customResource v1alpha1.ChartConfig) bool {
	_, reasonOk := customResource.Annotations[annotation.CordonReason]
	_, untilOk := customResource.Annotations[annotation.CordonUntilDate]

	if reasonOk && untilOk {
		return true
	} else {
		return false
	}
}

func Namespace(customResource v1alpha1.ChartConfig) string {
	return customResource.Spec.Chart.Namespace
}

func ReleaseName(customResource v1alpha1.ChartConfig) string {
	return customResource.Spec.Chart.Release
}

func ReleaseStatus(customResource v1alpha1.ChartConfig) string {
	return customResource.Status.ReleaseStatus
}
