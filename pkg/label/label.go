// Package label contains common Kubernetes object labels. These are defined in
// https://github.com/giantswarm/fmt/blob/master/kubernetes/annotations_and_labels.md.
package label

const (
	// App is a standard label for Kubernetes resources.
	App = "app"

	// AppKubernetesManagedBy label is used to identify the component managing
	// Kubernetes resources.
	AppKubernetesManagedBy = "app.kubernetes.io/managed-by"

	// HelmServiceNameValue is used to identify when resources are managed by
	// Helm
	HelmServiceNameValue = "Helm"
)
