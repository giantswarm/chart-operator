package chart

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/appr"
)

const (
	// Name is the identifier of the resource.
	Name = "chartv1"
)

// Config represents the configuration used to create a new chart resource.
type Config struct {
	// Dependencies.
	K8sClient  kubernetes.Interface
	ApprClient *appr.Client
	Logger     micrologger.Logger
}

// Resource implements the chart resource.
type Resource struct {
	// Dependencies.
	k8sClient  kubernetes.Interface
	apprClient *appr.Client
	logger     micrologger.Logger
}

// New creates a new configured chart resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.ApprClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.ApprClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	r := &Resource{
		// Dependencies.
		k8sClient:  config.K8sClient,
		apprClient: config.ApprClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
