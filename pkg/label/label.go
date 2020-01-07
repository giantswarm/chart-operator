// Package label contains common Kubernetes object labels. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package label

const (
	// ManagedBy is set for Kubernetes resources managed by the operator.
	ManagedBy = "giantswarm.io/managed-by"
)
