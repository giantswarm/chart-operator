package healthz

import (
	"github.com/giantswarm/k8shealthz"
	"github.com/giantswarm/microendpoint/service/healthz"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

// Config represents the configuration used to create a healthz service.
type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

// New creates a new configured healthz service.
func New(config Config) (*Service, error) {
	var err error

	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}

	var k8sService healthz.Service
	{
		c := k8shealthz.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		k8sService, err = k8shealthz.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		K8s: k8sService,
	}

	return newService, nil
}

// Service is the healthz service collection.
type Service struct {
	K8s healthz.Service
}
