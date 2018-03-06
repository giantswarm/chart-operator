// Package appr holds the client code required to interact with a CNR backend.
package appr

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

// Config represents the configuration used to create a appr client.
type Config struct {
	Logger  micrologger.Logger
	Address string
}

// Client knows how to talk with a CNR server.
type Client struct {
	logger  micrologger.Logger
	address string
}

// New creates a new configured appr client.
func New(config Config) (*Client, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	if config.Address == "" {
		return nil, microerror.Maskf(invalidConfigError, "address must not be empty")
	}

	newAppr := &Client{
		logger:  config.Logger,
		address: config.Address,
	}

	return newAppr, nil
}
