package middleware

import (
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/chart-operator/service"
)

// Config represents the configuration used to construct middleware.
type Config struct {
	// Dependencies
	Logger  micrologger.Logger
	Service *service.Service
}

// New creates a new configured middleware.
func New(config Config) (*Middleware, error) {
	return &Middleware{}, nil
}

// Middleware is middleware collection.
type Middleware struct {
}
