// +build k8srequired

package setup

import (
	"github.com/giantswarm/app/v5/pkg/crd"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"

	"github.com/giantswarm/chart-operator/v2/integration/env"
	"github.com/giantswarm/chart-operator/v2/integration/release"
)

type Config struct {
	CRDGetter  *crd.CRDGetter
	HelmClient helmclient.Interface
	K8s        *k8sclient.Setup
	K8sClients k8sclient.Interface
	Logger     micrologger.Logger
	Release    *release.Release
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

	var crdGetter *crd.CRDGetter
	{
		c := crd.Config{
			Logger: logger,
		}

		crdGetter, err = crd.NewCRDGetter(c)
	}

	var k8sClients *k8sclient.Clients
	{
		c := k8sclient.ClientsConfig{
			Logger: logger,

			KubeConfigPath: env.KubeConfigPath(),
		}

		k8sClients, err = k8sclient.NewClients(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var k8sSetup *k8sclient.Setup
	{
		c := k8sclient.SetupConfig{
			Clients: k8sClients,
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
			Fs:         fs,
			K8sClient:  k8sClients.K8sClient(),
			Logger:     logger,
			RestClient: k8sClients.RESTClient(),
			RestConfig: k8sClients.RESTConfig(),
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
		CRDGetter:  crdGetter,
		HelmClient: helmClient,
		K8s:        k8sSetup,
		K8sClients: k8sClients,
		Logger:     logger,
		Release:    newRelease,
	}

	return c, nil
}
