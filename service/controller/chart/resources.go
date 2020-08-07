package chart

import (
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/crud"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/service/controller/chart/resource/chartmigration"
	"github.com/giantswarm/chart-operator/service/controller/chart/resource/release"
	"github.com/giantswarm/chart-operator/service/controller/chart/resource/status"
	"github.com/giantswarm/chart-operator/service/controller/chart/resource/tillermigration"
)

type chartResourcesConfig struct {
	// Dependencies.
	Fs         afero.Fs
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	// Settings.
	K8sWaitTimeout  time.Duration
	MaxRollback     int
	TillerNamespace string
}

func newChartResources(config chartResourcesConfig) ([]resource.Interface, error) {
	var err error

	// Dependencies.
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.TillerNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TillerNamespace must not be empty", config)
	}

	var chartMigrationResource resource.Interface
	{
		c := chartmigration.Config{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		chartMigrationResource, err = chartmigration.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var releaseResource resource.Interface
	{
		c := release.Config{
			// Dependencies
			Fs:         config.Fs,
			G8sClient:  config.G8sClient,
			HelmClient: config.HelmClient,
			K8sClient:  config.K8sClient,
			Logger:     config.Logger,

			// Settings
			K8sWaitTimeout:  config.K8sWaitTimeout,
			MaxRollback:     config.MaxRollback,
			TillerNamespace: config.TillerNamespace,
		}

		ops, err := release.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		releaseResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var statusResource resource.Interface
	{
		c := status.Config{
			G8sClient:  config.G8sClient,
			HelmClient: config.HelmClient,
			Logger:     config.Logger,
		}

		statusResource, err = status.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tillerMigrationResource resource.Interface
	{
		c := tillermigration.Config{
			G8sClient: config.G8sClient,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			TillerNamespace: config.TillerNamespace,
		}

		tillerMigrationResource, err = tillermigration.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		chartMigrationResource,
		tillerMigrationResource,
		releaseResource,
		statusResource,
	}

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}
		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}

func toCRUDResource(logger micrologger.Logger, ops crud.Interface) (*crud.Resource, error) {
	c := crud.ResourceConfig{
		Logger: logger,
		CRUD:   ops,
	}

	r, err := crud.NewResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
