package chart

import (
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/chart-operator/pkg/project"
)

const chartControllerSuffix = "-chart"

type Config struct {
	Fs         afero.Fs
	HelmClient helmclient.Interface
	K8sClient  k8sclient.Interface
	Logger     micrologger.Logger
}

type Chart struct {
	*controller.Controller
}

func NewChart(config Config) (*Chart, error) {
	var err error

	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	resourceSets, err := newChartResourceSets(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var chartController *controller.Controller
	{
		c := controller.Config{
			CRD:          v1alpha1.NewChartCRD(),
			K8sClient:    config.K8sClient,
			Logger:       config.Logger,
			ResourceSets: resourceSets,
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.Chart)
			},

			Name:         project.Name() + chartControllerSuffix,
			ResyncPeriod: 5 * time.Minute,
		}

		chartController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Chart{
		Controller: chartController,
	}

	return c, nil
}

func newChartResourceSets(config Config) ([]*controller.ResourceSet, error) {
	var err error

	var resourceSet *controller.ResourceSet
	{
		c := chartResourceSetConfig{
			Fs:         config.Fs,
			G8sClient:  config.K8sClient.G8sClient(),
			HelmClient: config.HelmClient,
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,
		}

		resourceSet, err = newChartResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resourceSets := []*controller.ResourceSet{
		resourceSet,
	}

	return resourceSets, nil
}
