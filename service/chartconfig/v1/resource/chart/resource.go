package chart

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/appr"
	"github.com/giantswarm/chart-operator/service/chartconfig/v1/helm"
)

const (
	// Name is the identifier of the resource.
	Name = "chartv1"
)

// Config represents the configuration used to create a new chart resource.
type Config struct {
	// Dependencies.
	ApprClient appr.Interface
	HelmClient helm.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
}

// Resource implements the chart resource.
type Resource struct {
	// Dependencies.
	apprClient appr.Interface
	helmClient helm.Interface
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
}

// New creates a new configured chart resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.ApprClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ApprClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		apprClient: config.ApprClient,
		helmClient: config.HelmClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func toChartState(v interface{}) (ChartState, error) {
	if v == nil {
		return ChartState{}, nil
	}

	chartState, ok := v.(*ChartState)
	if !ok {
		return ChartState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", chartState, v)
	}

	return *chartState, nil
}
