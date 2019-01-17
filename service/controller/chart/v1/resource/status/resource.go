package status

import (
	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "statusv1"
)

// Config represents the configuration used to create a new status resource.
type Config struct {
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	Logger     micrologger.Logger
}

// Resource implements the status resource.
type Resource struct {
	g8sClient  versioned.Interface
	helmClient helmclient.Interface
	logger     micrologger.Logger
}

// New creates a new configured status resource.
func New(config Config) (*Resource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		g8sClient:  config.G8sClient,
		helmClient: config.HelmClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

// Equals asseses the equality of ChartStatuses with regards to distinguishing
// fields.
func Equals(a, b v1alpha1.ChartStatus) bool {
	if a.AppVersion != b.AppVersion {
		return false
	}
	if a.LastDeployed != b.LastDeployed {
		return false
	}
	if a.Status != b.Status {
		return false
	}
	if a.Version != b.Version {
		return false
	}

	return true
}

// IsEmpty checks if a ChartStatus is empty.
func IsEmpty(c v1alpha1.ChartStatus) bool {
	return Equals(c, v1alpha1.ChartStatus{})
}
