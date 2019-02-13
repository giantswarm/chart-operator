package setup

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/e2e-harness/pkg/framework/resource"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

const (
	namespace       = "giantswarm"
	tillerNamespace = "giantswarm"
)

type Config struct {
	Host       *framework.Host
	HelmClient *helmclient.Client
	Logger     micrologger.Logger
	Release    *release.Release
	Resource   *resource.Resource
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

	var host *framework.Host
	{
		c := framework.HostConfig{
			Logger: logger,

			ClusterID:       "n/a",
			VaultToken:      "n/a",
			TargetNamespace: namespace,
		}

		host, err = framework.NewHost(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var helmClient *helmclient.Client
	{
		c := helmclient.Config{
			Logger:          logger,
			K8sClient:       host.K8sClient(),
			RestConfig:      host.RestConfig(),
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
			ExtClient:  host.ExtClient(),
			G8sClient:  host.G8sClient(),
			HelmClient: helmClient,
			K8sClient:  host.K8sClient(),
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
		Host:       host,
		HelmClient: helmClient,
		Logger:     logger,
		Release:    newRelease,
		// Resource is deprecated and used by legacy chartconfig tests.
		Resource: newResource,
	}

	return c, nil
}
