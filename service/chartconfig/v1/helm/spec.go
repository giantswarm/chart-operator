package helm

import "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"

// Interface describes the methods provided by the helm client.
type Interface interface {
	GetReleaseContent(v1alpha1.ChartConfig) (*Release, error)
}

// Release returns information about a Helm release.
type Release struct {
	// Name is the name of the Helm release.
	Name string
	// Status is the Helm status code of the release.
	Status string
}
