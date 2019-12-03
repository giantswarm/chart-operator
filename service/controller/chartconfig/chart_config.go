package chartconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/chart-operator/pkg/project"
	v5 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v5"
	v6 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v6"
	v7 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v7"
)

const chartConfigControllerSuffix = "-chartconfig"

type Config struct {
	ApprClient apprclient.Interface
	Fs         afero.Fs
	HelmClient helmclient.Interface
	K8sClient  k8sclient.Interface
	Logger     micrologger.Logger

	WatchNamespace string
}

type ChartConfig struct {
	*controller.Controller
}

func NewChartConfig(config Config) (*ChartConfig, error) {
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

	var resourceSetV5 *controller.ResourceSet
	{
		c := v5.ResourceSetConfig{
			ApprClient: config.ApprClient,
			Fs:         config.Fs,
			G8sClient:  config.K8sClient.G8sClient(),
			HelmClient: config.HelmClient,
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,
		}

		resourceSetV5, err = v5.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV6 *controller.ResourceSet
	{
		c := v6.ResourceSetConfig{
			ApprClient: config.ApprClient,
			Fs:         config.Fs,
			G8sClient:  config.K8sClient.G8sClient(),
			HelmClient: config.HelmClient,
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,
		}

		resourceSetV6, err = v6.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV7 *controller.ResourceSet
	{
		c := v7.ResourceSetConfig{
			ApprClient: config.ApprClient,
			Fs:         config.Fs,
			G8sClient:  config.K8sClient.G8sClient(),
			HelmClient: config.HelmClient,
			K8sClient:  config.K8sClient.K8sClient(),
			Logger:     config.Logger,
		}

		resourceSetV7, err = v7.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartConfigController *controller.Controller
	{
		c := controller.Config{
			CRD:       v1alpha1.NewChartConfigCRD(),
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSetV5,
				resourceSetV6,
				resourceSetV7,
			},
			NewRuntimeObjectFunc: func() runtime.Object {
				return new(v1alpha1.ChartConfig)
			},

			Name: project.Name() + chartConfigControllerSuffix,
		}

		chartConfigController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &ChartConfig{
		Controller: chartConfigController,
	}

	return c, nil
}
