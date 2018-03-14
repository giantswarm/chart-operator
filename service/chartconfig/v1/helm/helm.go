package helm

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	helmclient "k8s.io/helm/pkg/helm"
)

const (
	connectionTimeoutSecs = 5
)

// Config represents the configuration used to create a helm client.
type Config struct {
	HelmClient Interface
	Logger     micrologger.Logger

	Host string
}

// Client knows how to talk with a Helm Tiller server.
type Client struct {
	helmClient *helmclient.Client
	logger     micrologger.Logger
}

// New creates a new configured Helm client.
func New(config Config) (*Client, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Host == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Host must not be empty", config)
	}

	hc := helmclient.NewClient(helmclient.Host(config.Host),
		helmclient.ConnectTimeout(connectionTimeoutSecs))

	newHelm := &Client{
		helmClient: hc,
		logger:     config.Logger,
	}

	return newHelm, nil
}

// GetReleaseContent gets the current status of the Helm Release.
func (c *Client) GetReleaseContent(customObject v1alpha1.ChartConfig) (*Release, error) {
	return nil, nil
}
