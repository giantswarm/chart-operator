package setup

import (
	"github.com/giantswarm/chart-operator/integration/env"
	"github.com/giantswarm/e2e-harness/pkg/framework/resource"
	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2esetup/k8s"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	namespace       = "giantswarm"
	tillerNamespace = "kube-system"
)

type Config struct {
	CPK8sClients *k8s.Clients
	CPK8sSetup   *k8s.Setup
	HelmClient   *helmclient.Client
	Logger       micrologger.Logger
	Release      *release.Release
	Resource     *resource.Resource
}

func NewConfig() (Config, error) {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{}
		logger, err = micrologger.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var cpK8sClients *k8s.Clients
	{
		kubeConfigPath := env.KubeConfigPath()
		if kubeConfigPath == "" {
			kubeConfigPath = harness.DefaultKubeConfig
		}

		c := k8s.ClientsConfig{
			Logger: logger,

			KubeConfigPath: kubeConfigPath,
		}

		cpK8sClients, err = k8s.NewClients(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var cpK8sSetup *k8s.Setup
	{
		c := k8s.SetupConfig{
			K8sClient: cpK8sClients.K8sClient(),
			Logger:    logger,
		}

		cpK8sSetup, err = k8s.NewSetup(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var helmClient *helmclient.Client
	{
		c := helmclient.Config{
			Logger:          logger,
			K8sClient:       cpK8sClients.K8sClient(),
			RestConfig:      cpK8sClients.RestConfig(),
			TillerNamespace: tillerNamespace,
		}
		helmClient, err = helmclient.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var newRelease *release.Release
	{
		c := release.Config{
			ExtClient:  cpK8sClients.ExtClient(),
			G8sClient:  cpK8sClients.G8sClient(),
			HelmClient: helmClient,
			K8sClient:  cpK8sClients.K8sClient(),
			Logger:     logger,

			Namespace: namespace,
		}

		newRelease, err = release.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var newResource *resource.Resource
	{
		c := resource.Config{
			Logger:     logger,
			HelmClient: helmClient,
			Namespace:  namespace,
		}
		newResource, err = resource.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	c := Config{
		CPK8sClients: cpK8sClients,
		CPK8sSetup:   cpK8sSetup,
		HelmClient:   helmClient,
		Logger:       logger,
		Release:      newRelease,
		// Resource is deprecated and used by legacy chartconfig tests.
		Resource: newResource,
	}

	return c, nil
}
