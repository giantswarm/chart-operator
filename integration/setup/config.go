// +build k8srequired

package setup

import (
	"github.com/giantswarm/helmclient"
	k8sclientv2 "github.com/giantswarm/k8sclient/v2/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v3/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"

	"github.com/giantswarm/chart-operator/integration/env"
	"github.com/giantswarm/chart-operator/integration/release"
)

type Config struct {
	HelmClient   helmclient.Interface
	K8s          *k8sclient.Setup
	K8sClientsV2 k8sclientv2.Interface
	K8sClients   k8sclient.Interface
	Logger       micrologger.Logger
	Release      *release.Release
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

	var cpK8sClients *k8sclient.Clients
	{
		c := k8sclient.ClientsConfig{
			Logger: logger,

			KubeConfigPath: env.KubeConfigPath(),
		}

		cpK8sClients, err = k8sclient.NewClients(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	// cpK8sClientsV2 is to create the chartconfig CRD which will not be
	// graduated from v1beta1 to v1 since its deprecated.
	var cpK8sClientsV2 *k8sclientv2.Clients
	{
		c := k8sclientv2.ClientsConfig{
			Logger: logger,

			KubeConfigPath: env.KubeConfigPath(),
		}

		cpK8sClientsV2, err = k8sclientv2.NewClients(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var k8sSetup *k8sclient.Setup
	{
		c := k8sclient.SetupConfig{
			Clients: cpK8sClients,
			Logger:  logger,
		}

		k8sSetup, err = k8sclient.NewSetup(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	fs := afero.NewOsFs()

	var helmClient *helmclient.Client
	{
		c := helmclient.Config{
			Fs:        fs,
			K8sClient: cpK8sClients,
			Logger:    logger,
		}
		helmClient, err = helmclient.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var newRelease *release.Release
	{
		c := release.Config{
			HelmClient: helmClient,
			Logger:     logger,
		}

		newRelease, err = release.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	c := Config{
		HelmClient:   helmClient,
		K8s:          k8sSetup,
		K8sClients:   cpK8sClients,
		K8sClientsV2: cpK8sClientsV2,
		Logger:       logger,
		Release:      newRelease,
	}

	return c, nil
}
