package chartconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/k8sclient/k8scrdclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/spf13/afero"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"

	v5 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v5"
	v6 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v6"
	v7 "github.com/giantswarm/chart-operator/service/controller/chartconfig/v7"
)

const chartConfigControllerSuffix = "-chartconfig"

type Config struct {
	ApprClient   apprclient.Interface
	Fs           afero.Fs
	G8sClient    versioned.Interface
	HelmClient   helmclient.Interface
	K8sClient    kubernetes.Interface
	K8sExtClient apiextensionsclient.Interface
	Logger       micrologger.Logger

	ProjectName    string
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
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.K8sExtClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sExtClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	var crdClient *k8scrdclient.CRDClient
	{
		c := k8scrdclient.Config{
			K8sExtClient: config.K8sExtClient,
			Logger:       config.Logger,
		}

		crdClient, err = k8scrdclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newInformer *informer.Informer
	{
		c := informer.Config{
			Logger:  config.Logger,
			Watcher: config.G8sClient.CoreV1alpha1().ChartConfigs(config.WatchNamespace),

			RateWait:     informer.DefaultRateWait,
			ResyncPeriod: informer.DefaultResyncPeriod,
		}

		newInformer, err = informer.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV5 *controller.ResourceSet
	{
		c := v5.ResourceSetConfig{
			ApprClient:  config.ApprClient,
			Fs:          config.Fs,
			G8sClient:   config.G8sClient,
			HelmClient:  config.HelmClient,
			K8sClient:   config.K8sClient,
			Logger:      config.Logger,
			ProjectName: config.ProjectName,
		}

		resourceSetV5, err = v5.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV6 *controller.ResourceSet
	{
		c := v6.ResourceSetConfig{
			ApprClient:  config.ApprClient,
			Fs:          config.Fs,
			G8sClient:   config.G8sClient,
			HelmClient:  config.HelmClient,
			K8sClient:   config.K8sClient,
			Logger:      config.Logger,
			ProjectName: config.ProjectName,
		}

		resourceSetV6, err = v6.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resourceSetV7 *controller.ResourceSet
	{
		c := v7.ResourceSetConfig{
			ApprClient:  config.ApprClient,
			Fs:          config.Fs,
			G8sClient:   config.G8sClient,
			HelmClient:  config.HelmClient,
			K8sClient:   config.K8sClient,
			Logger:      config.Logger,
			ProjectName: config.ProjectName,
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
			CRDClient: crdClient,
			Informer:  newInformer,
			Logger:    config.Logger,
			ResourceSets: []*controller.ResourceSet{
				resourceSetV5,
				resourceSetV6,
				resourceSetV7,
			},
			RESTClient: config.G8sClient.CoreV1alpha1().RESTClient(),

			Name: config.ProjectName + chartConfigControllerSuffix,
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
