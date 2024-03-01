package service

import (
	"context"
	"fmt"
	"sync"

	applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v7/pkg/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/giantswarm/chart-operator/v3/flag"
	"github.com/giantswarm/chart-operator/v3/pkg/project"
	"github.com/giantswarm/chart-operator/v3/service/collector"
	"github.com/giantswarm/chart-operator/v3/service/controller/chart"

	"github.com/giantswarm/chart-operator/v3/service/internal/clientpair"
)

const (
	publicClientSAName      = "automation"
	publicClientSANamespace = "default"
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
	var restConfigPrv *rest.Config
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

		restConfigPrv, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	restConfigPub := rest.CopyConfig(restConfigPrv)

	fs := afero.NewOsFs()

	// k8sPrvClient runs under the chart-operator default permissions and hence
	// has elevated privileges in the cluster. It is meant to be used for
	// reconciling giantswarm-protected namespaces.
	var k8sPrvClient k8sclient.Interface
	var prvHelmClient helmclient.Interface
	{
		k8sPrvClient, err = newK8sClient(config, restConfigPrv)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		prvHelmClient, err = newHelmClient(config, k8sPrvClient, fs)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// k8sPubClient runs under `default:automation` Service Account when using
	// split client configuration. This client is meant to be used for reconciling
	// customer namespaces. For Workload Clusters it is `nil` and only prvHelmClient
	// use used.
	var k8sPubClient k8sclient.Interface
	var pubHelmClient helmclient.Interface
	if config.Viper.GetBool(config.Flag.Service.Helm.SplitClient) {

		// Using public client should result in permissions error, like for example
		// Upgrade "hello-world" failed: failed to create resource:
		//   rolebindings.rbac.authorization.k8s.io is forbidden:
		//     User "system:serviceaccount:default:automation" cannot create resource
		//     "rolebindings" in API group "rbac.authorization.k8s.io" in the namespace
		//     "giantswarm".
		restConfigPub.Impersonate = rest.ImpersonationConfig{
			UserName: fmt.Sprintf(
				"system:serviceaccount:%s:%s",
				publicClientSANamespace,
				publicClientSAName,
			),
		}

		k8sPubClient, err = newK8sClient(config, restConfigPub)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		pubHelmClient, err = newHelmClient(config, k8sPubClient, fs)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	cpConfig := clientpair.ClientPairConfig{
		Logger: config.Logger,

		NamespaceWhitelist: config.Viper.GetStringSlice(config.Flag.Service.Helm.NamespaceWhitelist),

		PrvHelmClient: prvHelmClient,
		PubHelmClient: pubHelmClient,
	}

	helmClients, err := clientpair.NewClientPair(cpConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var chartController *chart.Chart
	{
		c := chart.Config{
			Fs:          fs,
			HelmClients: helmClients,
			Logger:      config.Logger,
			K8sClient:   k8sPrvClient,

			ResyncPeriod: config.Viper.GetDuration(config.Flag.Service.Controller.ResyncPeriod),

			HTTPClientTimeout: config.Viper.GetDuration(config.Flag.Service.Helm.HTTP.ClientTimeout),
			K8sWaitTimeout:    config.Viper.GetDuration(config.Flag.Service.Helm.Kubernetes.WaitTimeout),
			K8sWatchNamespace: config.Viper.GetString(config.Flag.Service.Kubernetes.Watch.Namespace),
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
			// Collector must use client with elevated privileges in order to
			// look for orphaned ConfigMap and Secrets
			K8sClient: k8sPrvClient,
			Logger:    config.Logger,

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

func newHelmClient(config Config, k8sClient k8sclient.Interface, fs afero.Fs) (*helmclient.Client, error) {
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

	helmClient, err := helmclient.New(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return helmClient, err
}

func newK8sClient(config Config, restConfig *rest.Config) (k8sclient.Interface, error) {
	c := k8sclient.ClientsConfig{
		Logger: config.Logger,
		SchemeBuilder: k8sclient.SchemeBuilder{
			applicationv1alpha1.AddToScheme,
		},

		RestConfig: restConfig,
	}

	k8sClient, err := k8sclient.NewClients(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return k8sClient, nil
}
