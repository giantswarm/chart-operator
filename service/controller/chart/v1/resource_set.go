package v1

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/resource"
	"github.com/giantswarm/operatorkit/resource/crud"
	"github.com/giantswarm/operatorkit/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/resource/wrapper/retryresource"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/controllercontext"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/resource/chartmigration"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/resource/release"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/resource/releasemigration"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/resource/status"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/resource/tiller"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/resource/tillermigration"
)

// ResourceSetConfig contains necessary dependencies and settings for
// Chart controller ResourceSet configuration.
type ResourceSetConfig struct {
	// Dependencies.
	Fs         afero.Fs
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	// Settings.
	HandledVersionBundles []string
	TillerNamespace       string
}

// NewResourceSet returns a configured Chart controller ResourceSet.
func NewResourceSet(config ResourceSetConfig) (*controller.ResourceSet, error) {
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

	var tillerResource resource.Interface
	{
		c := tiller.Config{
			HelmClient: config.HelmClient,
			Logger:     config.Logger,
		}

		tillerResource, err = tiller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var releaseMigrationResource resource.Interface
	{
		c := releasemigration.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			TillerNamespace: config.TillerNamespace,
		}

		releaseMigrationResource, err = releasemigration.New(c)
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
		tillerResource,
		tillerMigrationResource,
		releaseMigrationResource,
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

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		cc := controllercontext.Context{}
		ctx = controllercontext.NewContext(ctx, cc)

		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		cr, err := key.ToCustomResource(obj)
		if err != nil {
			return false
		}

		if key.VersionLabel(cr) == VersionBundle().Version {
			return true
		}

		return false
	}

	var resourceSet *controller.ResourceSet
	{
		c := controller.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = controller.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
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
