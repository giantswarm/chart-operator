package helm

import "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"

// Interface describes the methods provided by the helm client.
type Interface interface {
	GetReleaseContent(v1alpha1.ChartConfig) (*ReleaseContent, error)
	GetReleaseHistory(v1alpha1.ChartConfig) (*ReleaseHistory, error)
}

// ReleaseContent returns status information about a Helm Release.
type ReleaseContent struct {
	// Name is the name of the Helm Release.
	Name string
	// Status is the Helm status code of the Release.
	Status string
	// Values are the values provided when installing the Helm Release.
	Values map[string]interface{}
}

// ReleaseHistory returns version information about a Helm Release.
type ReleaseHistory struct {
	// Name is the name of the Helm Release.
	Name string
	// ReleaseVersion is the version of the Helm Chart to be deployed.
	ReleaseVersion string
}
