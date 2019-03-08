package helmclienttest

import (
	"context"

	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/helmclient"
)

type Config struct {
	DefaultError          error
	DefaultReleaseContent *helmclient.ReleaseContent
	DefaultReleaseHistory *helmclient.ReleaseHistory
	LoadChartError        error
	LoadChartResponse     helmclient.Chart
	PullChartTarballError error
	PullChartTarballPath  string
}

type Client struct {
	defaultError          error
	defaultReleaseContent *helmclient.ReleaseContent
	defaultReleaseHistory *helmclient.ReleaseHistory
	loadChartError        error
	loadChartResponse     helmclient.Chart
	pullChartTarballError error
	pullChartTarballPath  string
}

func New(config Config) helmclient.Interface {
	c := &Client{
		defaultError:          config.DefaultError,
		defaultReleaseContent: config.DefaultReleaseContent,
		defaultReleaseHistory: config.DefaultReleaseHistory,
		loadChartError:        config.LoadChartError,
		loadChartResponse:     config.LoadChartResponse,
		pullChartTarballError: config.PullChartTarballError,
		pullChartTarballPath:  config.PullChartTarballPath,
	}

	return c
}

func (c *Client) DeleteRelease(ctx context.Context, releaseName string, options ...helm.DeleteOption) error {
	if c.defaultError != nil {
		return c.defaultError
	}

	return nil
}

func (c *Client) EnsureTillerInstalled(ctx context.Context) error {
	return nil
}

func (c *Client) EnsureTillerInstalledWithValues(ctx context.Context, values []string) error {
	return nil
}

func (c *Client) GetReleaseContent(ctx context.Context, releaseName string) (*helmclient.ReleaseContent, error) {
	if c.defaultError != nil {
		return nil, c.defaultError
	}

	return c.defaultReleaseContent, nil
}

func (c *Client) GetReleaseHistory(ctx context.Context, releaseName string) (*helmclient.ReleaseHistory, error) {
	if c.defaultError != nil {
		return nil, c.defaultError
	}

	return c.defaultReleaseHistory, nil
}

func (c *Client) InstallReleaseFromTarball(ctx context.Context, path, ns string, options ...helm.InstallOption) error {
	return nil
}

func (c *Client) ListReleaseContents(ctx context.Context) ([]*helmclient.ReleaseContent, error) {
	return nil, nil
}

func (c *Client) LoadChart(ctx context.Context, chartPath string) (helmclient.Chart, error) {
	if c.loadChartError != nil {
		return helmclient.Chart{}, c.loadChartError
	}

	return c.loadChartResponse, nil
}

func (c *Client) PingTiller(ctx context.Context) error {
	return nil
}

func (c *Client) PullChartTarball(ctx context.Context, tarballURL string) (string, error) {
	if c.pullChartTarballError != nil {
		return "", c.pullChartTarballError
	}

	return c.pullChartTarballPath, nil
}

func (c *Client) RunReleaseTest(ctx context.Context, releaseName string, options ...helm.ReleaseTestOption) error {
	return nil
}

func (c *Client) UpdateReleaseFromTarball(ctx context.Context, releaseName, path string, options ...helm.UpdateOption) error {
	return nil
}
