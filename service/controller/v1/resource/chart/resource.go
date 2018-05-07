package chart

import (
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "chartv1"
)

// Config represents the configuration used to create a new chart resource.
type Config struct {
	// Dependencies.
	ApprClient apprclient.Interface
	Fs         afero.Fs
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
}

// Resource implements the chart resource.
type Resource struct {
	// Dependencies.
	apprClient apprclient.Interface
	fs         afero.Fs
	helmClient helmclient.Interface
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
}

// New creates a new configured chart resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.ApprClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ApprClient must not be empty", config)
	}
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Ensure Tiller is installed in the configured namespace.
	err := config.HelmClient.EnsureTillerInstalled()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r := &Resource{
		// Dependencies.
		apprClient: config.ApprClient,
		fs:         config.Fs,
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
