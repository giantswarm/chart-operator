package v5

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/giantswarm/operatorkit/controller/resource/metricsresource"
	"github.com/giantswarm/operatorkit/controller/resource/retryresource"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v5/key"
	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v5/resource/chart"
	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v5/resource/chartstatus"
	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v5/resource/tiller"
)

// ResourceSetConfig contains necessary dependencies and settings for
// ChartConfig controller ResourceSet configuration.
type ResourceSetConfig struct {
	// Dependencies.
	ApprClient apprclient.Interface
	Fs         afero.Fs
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	// Settings.
	HandledVersionBundles []string
	ProjectName           string
}

// NewResourceSet returns a configured ChartConfig controller ResourceSet.
func NewResourceSet(config ResourceSetConfig) (*controller.ResourceSet, error) {
	var err error

	// Dependencies.
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Settings.
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ProjectName must not be empty", config)
	}

	var chartResource controller.Resource
	{
		c := chart.Config{
			ApprClient: config.ApprClient,
			Fs:         config.Fs,
			HelmClient: config.HelmClient,
			K8sClient:  config.K8sClient,
			Logger:     config.Logger,
		}

		ops, err := chart.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		chartResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var chartStatusResource controller.Resource
	{
		c := chartstatus.Config{
			G8sClient:  config.G8sClient,
			HelmClient: config.HelmClient,
			Logger:     config.Logger,
		}

		chartStatusResource, err = chartstatus.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tillerResource controller.Resource
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

	resources := []controller.Resource{
		tillerResource,
		chartResource,
		chartStatusResource,
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
		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		chartConfig, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		if key.VersionBundleVersion(chartConfig) == VersionBundle().Version {
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

func toCRUDResource(logger micrologger.Logger, ops controller.CRUDResourceOps) (*controller.CRUDResource, error) {
	c := controller.CRUDResourceConfig{
		Logger: logger,
		Ops:    ops,
	}

	r, err := controller.NewCRUDResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
