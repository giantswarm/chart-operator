package chart

import (
	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/chart-operator/pkg/project"
	v1 "github.com/giantswarm/chart-operator/service/controller/chart/v1"
)

const chartConfigControllerSuffix = "-chart"

type Config struct {
	Fs         afero.Fs
	HelmClient helmclient.Interface
	K8sClient  k8sclient.Interface
	Logger     micrologger.Logger

	WatchNamespace string
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

	var resourceSetV1 *controller.ResourceSet
	{
		c := v1.ResourceSetConfig{
			Fs:         config.Fs,
			G8sClient:  config.K8sClient.G8sClient(),
			HelmClient: config.HelmClient,
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,
		}

		resourceSetV1, err = v1.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartController *controller.Controller
	{
		c := controller.Config{
			CRD:       v1alpha1.NewChartCRD(),
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSetV1,
			},
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.Chart)
			},

			Name: project.Name() + chartConfigControllerSuffix,
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
