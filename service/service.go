package service

import (
	"context"
	"sync"
	"time"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/k8sclient/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"github.com/giantswarm/versionbundle"

	"github.com/giantswarm/chart-operator/flag"
	"github.com/giantswarm/chart-operator/pkg/project"
	"github.com/giantswarm/chart-operator/service/collector"
	"github.com/giantswarm/chart-operator/service/controller/chart"
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
				corev1alpha1.AddToScheme,
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
		c := helmclient.Config{
			K8sClient: k8sClient.K8sClient(),
			Logger:    config.Logger,

			EnsureTillerInstalledMaxWait: 30 * time.Second,
			HTTPClientTimeout:            config.Viper.GetDuration(config.Flag.Service.Helm.HTTP.ClientTimeout),
			RestConfig:                   restConfig,
			TillerImageRegistry:          config.Viper.GetString(config.Flag.Service.Image.Registry),
			TillerNamespace:              config.Viper.GetString(config.Flag.Service.Helm.TillerNamespace),
			TillerUpgradeEnabled:         true,
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
			Description:    project.Description(),
			GitCommit:      project.GitSHA(),
			Name:           project.Name(),
			Source:         project.Source(),
			Version:        project.Version(),
			VersionBundles: []versionbundle.Bundle{project.NewVersionBundle()},
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
		go s.operatorCollector.Boot(ctx)

		go s.chartController.Boot(ctx)
	})
}
