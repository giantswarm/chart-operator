package service

import (
	"context"
	"sync"

	applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/k8sclient/v6/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v6/pkg/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/giantswarm/chart-operator/v2/flag"
	"github.com/giantswarm/chart-operator/v2/pkg/project"
	"github.com/giantswarm/chart-operator/v2/service/collector"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	Logger micrologger.Logger

	// Settings.
	Flag  *flag.Flag
	Viper *viper.Viper
}

// Service is a type providing implementation of microkit service interface.
type Service struct {
	Version *version.Service

	// Internals
	bootOnce          sync.Once
	chartController   *chart.Chart
	operatorCollector *collector.Set
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

			Address:    config.Viper.GetString(config.Flag.Service.Kubernetes.Address),
			InCluster:  config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			KubeConfig: config.Viper.GetString(config.Flag.Service.Kubernetes.KubeConfig),
			TLS: k8srestconfig.ConfigTLS{
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

	var k8sClient k8sclient.Interface
	{
		c := k8sclient.ClientsConfig{
			Logger: config.Logger,
			SchemeBuilder: k8sclient.SchemeBuilder{
				applicationv1alpha1.AddToScheme,
			},

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	fs := afero.NewOsFs()

	var helmClient *helmclient.Client
	{
		restMapper, err := apiutil.NewDynamicRESTMapper(rest.CopyConfig(k8sClient.RESTConfig()))
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := helmclient.Config{
			Fs:         fs,
			K8sClient:  k8sClient.K8sClient(),
			Logger:     config.Logger,
			RestClient: k8sClient.RESTClient(),
			RestConfig: k8sClient.RESTConfig(),
			RestMapper: restMapper,

			HTTPClientTimeout: config.Viper.GetDuration(config.Flag.Service.Helm.HTTP.ClientTimeout),
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartController *chart.Chart
	{
		c := chart.Config{
			Fs:         fs,
			HelmClient: helmClient,
			Logger:     config.Logger,
			K8sClient:  k8sClient,

			HTTPClientTimeout: config.Viper.GetDuration(config.Flag.Service.Helm.HTTP.ClientTimeout),
			K8sWaitTimeout:    config.Viper.GetDuration(config.Flag.Service.Helm.Kubernetes.WaitTimeout),
			MaxRollback:       config.Viper.GetInt(config.Flag.Service.Helm.MaxRollback),
			TillerNamespace:   config.Viper.GetString(config.Flag.Service.Helm.TillerNamespace),
		}

		chartController, err = chart.NewChart(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorCollector *collector.Set
	{
		c := collector.SetConfig{
			HelmClient: helmClient,
			K8sClient:  k8sClient,
			Logger:     config.Logger,

			TillerNamespace: config.Viper.GetString(config.Flag.Service.Helm.TillerNamespace),
		}

		operatorCollector, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		versionConfig := version.Config{
			Description: project.Description(),
			GitCommit:   project.GitSHA(),
			Name:        project.Name(),
			Source:      project.Source(),
			Version:     project.Version(),
		}

		versionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		bootOnce:          sync.Once{},
		chartController:   chartController,
		operatorCollector: operatorCollector,
	}

	return s, nil
}

// Boot starts top level service implementation.
func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go func() {
			err := s.operatorCollector.Boot(ctx)
			if err != nil {
				panic(microerror.JSON(err))
			}
		}()

		go s.chartController.Boot(ctx)
	})
}
