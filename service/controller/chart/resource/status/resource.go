package status

import (
	"time"

	"github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v2/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient/v2/pkg/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	Name = "status"

	// defaultHTTPClientTimeout is the timeout when updating app status.
	defaultHTTPClientTimeout = 5

	releaseStatusCordoned = "CORDONED"
)

// Config represents the configuration used to create a new status resource.
type Config struct {
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	Logger     micrologger.Logger

	HTTPClientTimeout time.Duration
}

// Resource implements the status resource.
type Resource struct {
	g8sClient  versioned.Interface
	helmClient helmclient.Interface
	logger     micrologger.Logger

	httpClientTimeout time.Duration
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

	if config.HTTPClientTimeout == 0 {
		config.HTTPClientTimeout = defaultHTTPClientTimeout
	}

	r := &Resource{
		g8sClient:  config.G8sClient,
		helmClient: config.HelmClient,
		logger:     config.Logger,

		httpClientTimeout: config.HTTPClientTimeout,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

// equals asseses the equality of ChartStatuses with regards to distinguishing
// fields.
func equals(a, b v1alpha1.ChartStatus) bool {
	if a.AppVersion != b.AppVersion {
		return false
	}
	if a.Release.LastDeployed != b.Release.LastDeployed {
		return false
	}
	if a.Reason != b.Reason {
		return false
	}
	if a.Release.Revision != b.Release.Revision {
		return false
	}
	if a.Release.Status != b.Release.Status {
		return false
	}
	if a.Version != b.Version {
		return false
	}

	return true
}
