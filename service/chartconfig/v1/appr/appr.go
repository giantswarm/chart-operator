// Package appr holds the client code required to interact with a CNR backend.
package appr

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
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

	base *url.URL
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

	u, err := url.Parse(config.Address + "/" + config.Organization + "/")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	newAppr := &Client{
		logger: config.Logger,

		base:       u,
		httpClient: hc,
	}

	return newAppr, nil
}

// DefaultRelease queries CNR for the default release name of the chart
// represented by the given custom object.
func (c *Client) DefaultRelease(customObject v1alpha1.ChartConfig) (string, error) {
	chartName := key.ChartName(customObject)

	req, err := c.newRequest("GET", chartName)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var pkg Package
	_, err = c.do(req, &pkg)

	if err != nil {
		return "", microerror.Mask(err)
	}

	return pkg.Release, nil
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
