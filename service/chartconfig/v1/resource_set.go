package v1

import (
	"context"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/key"
	"github.com/giantswarm/chart-operator/service/chartconfig/v1/resource/chart"
)

const (
	// ResourceRetries presents number of retries for failed Resource
	// operation before giving up.
	ResourceRetries uint64 = 3
)

// ResourceSetConfig contains necessary dependencies and settings for
// ChartConfig framework ResourceSet configuration.
type ResourceSetConfig struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	HandledVersionBundles []string
	ProjectName           string
}

// NewResourceSet returns a configured ChartConfig framework ResourceSet.
func NewResourceSet(config ResourceSetConfig) (*framework.ResourceSet, error) {
	var err error

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.ProjectName must not be empty")
	}

	var chartResource framework.Resource
	{
		c := chart.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		chartResource, err = chart.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []framework.Resource{
		chartResource,
	}

	// Wrap resources with retry and metrics.
	{
		retryWrapConfig := retryresource.WrapConfig{}

		retryWrapConfig.BackOffFactory = func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) }
		retryWrapConfig.Logger = config.Logger

		resources, err = retryresource.Wrap(resources, retryWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		metricsWrapConfig := metricsresource.WrapConfig{}

		metricsWrapConfig.Name = config.ProjectName

		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	initCtxFunc := func(ctx context.Context, obj interface{}) (context.Context, error) {
		return ctx, nil
	}

	handlesFunc := func(obj interface{}) bool {
		_, err := key.ToCustomObject(obj)
		if err != nil {
			return false
		}

		// Currently there is only one version to be handled. As long as the
		// object is of right type, it's good to go.

		return true
	}

	var resourceSet *framework.ResourceSet
	{
		c := framework.ResourceSetConfig{
			Handles:   handlesFunc,
			InitCtx:   initCtxFunc,
			Logger:    config.Logger,
			Resources: resources,
		}

		resourceSet, err = framework.NewResourceSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resourceSet, nil
}
