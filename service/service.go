package service

import (
	"context"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8srestconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/chart-operator/flag"
	"github.com/giantswarm/chart-operator/service/collector"
	"github.com/giantswarm/chart-operator/service/controller/chartconfig"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	HelmClient helmclient.Interface
	Logger     micrologger.Logger

	// Settings.
	Flag  *flag.Flag
	Viper *viper.Viper

	Description string
	GitCommit   string
	ProjectName string
	Source      string
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	Version *version.Service

	// Internals
	bootOnce              sync.Once
	chartConfigController *chartconfig.ChartConfig
	helmClient            helmclient.Interface
	metricsCollector      *collector.Collector
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

	fs := afero.NewOsFs()
	var apprClient *apprclient.Client
	{
		c := apprclient.Config{
			Fs:     fs,
			Logger: config.Logger,

			Address:      config.Viper.GetString(config.Flag.Service.CNR.Address),
			Organization: config.Viper.GetString(config.Flag.Service.CNR.Organization),
		}

		apprClient, err = apprclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var helmClient helmclient.Interface
	{
		c := helmclient.Config{
			K8sClient: k8sClient,
			Logger:    config.Logger,

			RestConfig:      restConfig,
			TillerNamespace: config.Viper.GetString(config.Flag.Service.Helm.TillerNamespace),
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var metricsCollector *collector.Collector
	{
		c := collector.Config{
			G8sClient:  g8sClient,
			HelmClient: helmClient,
			Logger:     config.Logger,
		}

		metricsCollector, err = collector.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartConfigController *chartconfig.ChartConfig
	{
		c := chartconfig.Config{
			ApprClient:   apprClient,
			Fs:           fs,
			HelmClient:   helmClient,
			G8sClient:    g8sClient,
			Logger:       config.Logger,
			K8sClient:    k8sClient,
			K8sExtClient: k8sExtClient,

			ProjectName:    config.ProjectName,
			WatchNamespace: config.Viper.GetString(config.Flag.Service.Kubernetes.Watch.Namespace),
		}

		chartConfigController, err = chartconfig.NewChartConfig(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		versionConfig := version.Config{
			Description:    config.Description,
			GitCommit:      config.GitCommit,
			Name:           config.ProjectName,
			Source:         config.Source,
			VersionBundles: NewVersionBundles(),
		}

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		bootOnce:              sync.Once{},
		chartConfigController: chartConfigController,
		helmClient:            helmClient,
		metricsCollector:      metricsCollector,
	}

	return s, nil
}

// Boot starts top level service implementation.
func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		prometheus.MustRegister(s.metricsCollector)
		s.helmClient.EnsureTillerInstalled(ctx)

		// Start the controller.
		go s.chartConfigController.Boot()
	})
}
