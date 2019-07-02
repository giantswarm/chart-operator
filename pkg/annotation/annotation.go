// Package annotation contains common Kubernetes metadata. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package annotation

const (
	// CordonReason is the name of the annotation that indicates
	// the reason of why chart-operator should not apply any update on this chart CR.
	CordonReason = "chart-operator.giantswarm.io/cordon-reason"

	// CordonUntilDate is the name of the annotation that indicates
	// the expiration date of rule of this cordon.
	CordonUntilDate = "chart-operator.giantswarm.io/cordon-until"
)
