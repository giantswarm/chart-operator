package service

import (
	"sync"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/chart-operator/flag"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Description string
	Flag        *flag.Flag
	GitCommit   string
	Name        string
	Source      string
	Viper       *viper.Viper
}

// New creates a new service with given configuration.
func New(config Config) (*Service, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	newService := &Service{
		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	// Internals
	bootOnce sync.Once
}

// Boot starts top level service implementation.
func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		// Insert service startup logic here.
	})
}
