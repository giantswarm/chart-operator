package endpoint

import (
	"github.com/giantswarm/microendpoint/endpoint/healthz"
	healthzservice "github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/chart-operator/server/middleware"
	"github.com/giantswarm/chart-operator/service"
)

// Config represents the configuration used to construct an endpoint.
type Config struct {
	// Dependencies
	Logger     micrologger.Logger
	Middleware *middleware.Middleware
	Service    *service.Service
}

// New creates a new endpoint with given configuration.
func New(config Config) (*Endpoint, error) {
	var err error

	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	if config.Service == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Service or it's Healthz descendents must not be empty")
	}

	var healthzEndpoint *healthz.Endpoint
	{
		c := healthz.DefaultConfig()
		c.Logger = config.Logger
		c.Services = []healthzservice.Service{
			config.Service.Healthz.K8s,
		}

		healthzEndpoint, err = healthz.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	endpoint := &Endpoint{
		Healthz: healthzEndpoint,
	}

	return endpoint, nil
}

// Endpoint is the endpoint collection.
type Endpoint struct {
	Healthz *healthz.Endpoint
}
