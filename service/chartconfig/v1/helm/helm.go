package helm

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/helm/pkg/chartutil"
	helmclient "k8s.io/helm/pkg/helm"
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

// DeleteRelease uninstalls a chart given its release name
func (c *Client) DeleteRelease(releaseName string, options ...helmclient.DeleteOption) error {
	_, err := c.helmClient.DeleteRelease(releaseName, options...)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// GetReleaseContent gets the current status of the Helm Release including any
// values provided when the chart was installed. The releaseName is the name
// of the Helm Release that is set when the Helm Chart is installed.
func (c *Client) GetReleaseContent(releaseName string) (*ReleaseContent, error) {
	resp, err := c.helmClient.ReleaseContent(releaseName)
	if IsReleaseNotFound(err) {
		return nil, microerror.Maskf(releaseNotFoundError, releaseName)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	// Raw values are returned by the Tiller API and we convert these to a map.
	raw := []byte(resp.Release.Config.Raw)
	values, err := chartutil.ReadValues(raw)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	content := &ReleaseContent{
		Name:   resp.Release.Name,
		Status: resp.Release.Info.Status.Code.String(),
		Values: values.AsMap(),
	}

	return content, nil
}

// GetReleaseHistory gets the current installed version of the Helm Release.
// The releaseName is the name of the Helm Release that is set when the Helm
// Chart is installed.
func (c *Client) GetReleaseHistory(releaseName string) (*ReleaseHistory, error) {
	var version string

	resp, err := c.helmClient.ReleaseHistory(releaseName, helmclient.WithMaxHistory(1))
	if IsReleaseNotFound(err) {
		return nil, microerror.Maskf(releaseNotFoundError, releaseName)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}
	if len(resp.Releases) > 1 {
		return nil, microerror.Maskf(tooManyResultsError, "%d releases found, expected 1", len(resp.Releases))
	}

	release := resp.Releases[0]
	if release.Chart != nil && release.Chart.Metadata != nil {
		version = release.Chart.Metadata.Version
	}

	history := &ReleaseHistory{
		Name:    release.Name,
		Version: version,
	}

	return history, nil
}

// InstallFromTarball installs a chart packaged in the given tarball.
func (c *Client) InstallFromTarball(path, ns string, options ...helmclient.InstallOption) error {
	_, err := c.helmClient.InstallRelease(path, ns, options...)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
