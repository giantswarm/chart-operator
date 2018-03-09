package helm

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

// Config represents the configuration used to create a helm client.
type Config struct {
	Logger micrologger.Logger
}

// Client knows how to talk with a Helm Tiller server.
type Client struct {
	logger micrologger.Logger
}

// New creates a new configured helm client.
func New(config Config) (*Client, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.logger must not be empty", config)
	}

	newHelm := &Client{
		logger: config.Logger,
	}

	return newHelm, nil
}
