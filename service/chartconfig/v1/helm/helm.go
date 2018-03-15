package helm

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/helm/pkg/chartutil"
	helmclient "k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/key"
)

const (
	connectionTimeoutSecs = 5
)

// Config represents the configuration used to create a helm client.
type Config struct {
	Logger micrologger.Logger

	Host string
}

// Client knows how to talk with a Helm Tiller server.
type Client struct {
	helmClient helmclient.Interface
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

// GetReleaseContent gets the current status of the Helm Release including any
// values provided when the chart was installed.
func (c *Client) GetReleaseContent(customObject v1alpha1.ChartConfig) (*Release, error) {
	releaseName := key.ReleaseName(customObject)

	resp, err := c.helmClient.ReleaseContent(releaseName)
	if err != nil && IsReleaseNotFound(err) {
		return nil, microerror.Maskf(releaseNotFoundError, releaseName)
	}
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Raw values are returned by the Tiller API and we convert these to a map.
	raw := []byte(resp.Release.Config.Raw)
	values, err := chartutil.ReadValues(raw)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	release := &Release{
		Name:   resp.Release.Name,
		Status: resp.Release.Info.Status.Code.String(),
		Values: values.AsMap(),
	}

	return release, nil
}
