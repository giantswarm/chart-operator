package helm

// Interface describes the methods provided by the helm client.
type Interface interface {
	GetReleaseContent(releaseName string) (*ReleaseContent, error)
	GetReleaseHistory(releaseName string) (*ReleaseHistory, error)
	InstallFromTarball(path string) error
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
	// Version is the version of the Helm Chart that has been deployed.
	Version string
}
