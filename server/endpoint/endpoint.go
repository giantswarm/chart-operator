package endpoint

import (
	"github.com/giantswarm/microendpoint/endpoint/healthz"
	healthzservice "github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/chart-operator/service"
)

// Config represents the configuration used to construct an endpoint.
type Config struct {
	Logger  micrologger.Logger
	Service *service.Service
}

// New creates a new endpoint with given configuration.
func New(config Config) (*Endpoint, error) {
	var err error

	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Service == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Service or it's Healthz descendents must not be empty", config)
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
