package service

import (
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/chart-operator/flag"
	"github.com/giantswarm/chart-operator/service/chartconfig"
	"github.com/giantswarm/chart-operator/service/chartconfig/v1/appr"
	"github.com/giantswarm/chart-operator/service/chartconfig/v1/helm"
	"github.com/giantswarm/chart-operator/service/healthz"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.
	Flag  *flag.Flag
	Viper *viper.Viper

	Description string
	GitCommit   string
	ProjectName string
	Source      string
}

// New creates a new service with given configuration.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Settings.
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Flag must not be empty", config)
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Viper must not be empty", config)
	}

	var err error

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: config.Logger,

			Address:   config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			InCluster: config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			TLS: k8srestconfig.TLSClientConfig{
				CAFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
				CrtFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
				KeyFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
			},
		}

		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sExtClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var apprClient *appr.Client
	{
		c := appr.Config{
			Logger: config.Logger,

			Address:      config.Viper.GetString(config.Flag.Service.CNR.Address),
			Organization: config.Viper.GetString(config.Flag.Service.CNR.Organization),
		}
		apprClient, err = appr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var helmClient *helm.Client
	{
		c := helm.Config{
			Logger: config.Logger,

			Host: config.Viper.GetString(config.Flag.Service.Helm.Host),
		}
		helmClient, err = helm.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var healthzService *healthz.Service
	{
		c := healthz.Config{
			K8sClient: k8sClient,
			Logger:    config.Logger,
		}

		healthzService, err = healthz.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartFramework *framework.Framework
	{
		c := chartconfig.ChartFrameworkConfig{
			ApprClient:   apprClient,
			HelmClient:   helmClient,
			G8sClient:    g8sClient,
			Logger:       config.Logger,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,

			ProjectName: config.ProjectName,
		}

		chartFramework, err = chartconfig.NewChartFramework(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		ChartFramework: chartFramework,
		Healthz:        healthzService,

		// Internals
		bootOnce: sync.Once{},
	}

	return newService, nil
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	ChartFramework *framework.Framework
	Healthz        *healthz.Service

	// Internals
	bootOnce sync.Once
}

// Boot starts top level service implementation.
func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		// Start the framework.
		go s.ChartFramework.Boot()
	})
}
