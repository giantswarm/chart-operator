// Package appr holds the client code required to interact with a CNR backend.
package appr

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/chart-operator/service/chartconfig/v1/key"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

// Config represents the configuration used to create a appr client.
type Config struct {
	Logger micrologger.Logger

	Address      string
	Organization string
}

// Client knows how to talk with a CNR server.
type Client struct {
	httpClient *http.Client
	logger     micrologger.Logger

	base         *url.URL
	organization string
}

// New creates a new configured appr client.
func New(config Config) (*Client, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}
	if config.Address == "" {
		return nil, microerror.Maskf(invalidConfigError, "address must not be empty")
	}
	if config.Organization == "" {
		return nil, microerror.Maskf(invalidConfigError, "organization must not be empty")
	}

	// set client timeout to prevent leakages.
	hc := &http.Client{
		Timeout: time.Second * httpClientTimeout,
	}

	u, err := url.Parse(config.Address + "/cnr/api/v1/")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	newAppr := &Client{
		logger: config.Logger,

		base:         u,
		httpClient:   hc,
		organization: config.Organization,
	}

	return newAppr, nil
}

// GetReleaseVersion queries CNR for the release version of the chart
// represented by the given custom object (including channel info).
func (c *Client) GetReleaseVersion(customObject v1alpha1.ChartConfig) (string, error) {
	chartName := key.ChartName(customObject)
	channelName := key.ChannelName(customObject)

	p := path.Join("packages", c.organization, chartName, "channels", channelName)

	req, err := c.newRequest("GET", p)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var ch Channel
	_, err = c.do(req, &ch)

	if err != nil {
		return "", microerror.Mask(err)
	}

	return ch.Current, nil
}

func (c *Client) newRequest(method, path string) (*http.Request, error) {
	u := &url.URL{Path: path}
	dest := c.base.ResolveReference(u)

	var buf io.ReadWriter

	req, err := http.NewRequest(method, dest.String(), buf)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return resp, nil
}
