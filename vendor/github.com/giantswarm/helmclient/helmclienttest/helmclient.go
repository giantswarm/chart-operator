package helmclienttest

import (
	"context"

	"github.com/giantswarm/helmclient"
	"k8s.io/helm/pkg/helm"
)

type Config struct {
	DefaultError          error
	DefaultHelmChart      *helmclient.Chart
	DefaultReleaseContent *helmclient.ReleaseContent
	DefaultReleaseHistory *helmclient.ReleaseHistory
	DefaultTarballPath    string
}

type Client struct {
	defaultError          error
	defaultHelmChart      *helmclient.Chart
	defaultReleaseContent *helmclient.ReleaseContent
	defaultReleaseHistory *helmclient.ReleaseHistory
	defaultTarballPath    string
}

func New(config Config) (helmclient.Interface, error) {
	c := &Client{
		defaultError:          config.DefaultError,
		defaultHelmChart:      config.DefaultHelmChart,
		defaultReleaseContent: config.DefaultReleaseContent,
		defaultReleaseHistory: config.DefaultReleaseHistory,
		defaultTarballPath:    config.DefaultTarballPath,
	}

	return c, nil
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

func (c *Client) LoadChart(ctx context.Context, chartPath string) (*helmclient.Chart, error) {
	if c.defaultError != nil {
		return nil, c.defaultError
	}

	return c.defaultHelmChart, nil
}

func (c *Client) PingTiller(ctx context.Context) error {
	return nil
}

func (c *Client) PullChartTarball(ctx context.Context, tarballURL string) (string, error) {
	if c.defaultError != nil {
		return "", c.defaultError
	}

	return c.defaultTarballPath, nil
}

func (c *Client) RunReleaseTest(ctx context.Context, releaseName string, options ...helm.ReleaseTestOption) error {
	return nil
}

func (c *Client) UpdateReleaseFromTarball(ctx context.Context, releaseName, path string, options ...helm.UpdateOption) error {
	return nil
}
