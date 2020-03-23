package apprclienttest

import (
	"context"
	"fmt"

	"github.com/giantswarm/apprclient"
)

type Config struct {
	DefaultError          error
	DefaultReleaseVersion string
}

type Client struct {
	defaultError          error
	defaultReleaseVersion string
}

func New(config Config) apprclient.Interface {
	c := &Client{
		defaultError:          config.DefaultError,
		defaultReleaseVersion: config.DefaultReleaseVersion,
	}

	return c
}

func (c *Client) DeleteRelease(ctx context.Context, name, release string) error {
	return nil
}

func (c *Client) GetReleaseVersion(ctx context.Context, name, channel string) (string, error) {
	if c.defaultError != nil {
		return "", fmt.Errorf("error getting default release")
	}

	return c.defaultReleaseVersion, nil
}

func (c *Client) PromoteChart(ctx context.Context, name, release, channel string) error {
	return nil
}

func (c *Client) PullChartTarball(ctx context.Context, name, channel string) (string, error) {
	return "", nil
}

func (c *Client) PullChartTarballFromRelease(ctx context.Context, name, release string) (string, error) {
	return "", nil
}

func (c *Client) PushChartTarball(ctx context.Context, name, release, tarballPath string) error {
	return nil
}
