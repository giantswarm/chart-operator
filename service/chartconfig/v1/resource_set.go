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

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/appr"
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
	// Dependencies.
	K8sClient  kubernetes.Interface
	ApprClient *appr.Client
	Logger     micrologger.Logger

	// Settings.
	HandledVersionBundles []string
	ProjectName           string
}

// NewResourceSet returns a configured ChartConfig framework ResourceSet.
func NewResourceSet(config ResourceSetConfig) (*framework.ResourceSet, error) {
	var err error

	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.ApprClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.ApprClient must not be empty")
	}

	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	// Settings.
	if config.ProjectName == "" {
		return nil, microerror.Maskf(invalidConfigError, "config.ProjectName must not be empty")
	}

	var chartResource framework.Resource
	{
		c := chart.Config{
			K8sClient:  config.K8sClient,
			ApprClient: config.ApprClient,
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

	resources := []framework.Resource{
		chartResource,
	}

	{
		c := retryresource.WrapConfig{
			BackOffFactory: func() backoff.BackOff { return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries) },
			Logger:         config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{
			Name: config.ProjectName,
		}

		resources, err = metricsresource.Wrap(resources, c)
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

func toCRUDResource(logger micrologger.Logger, ops framework.CRUDResourceOps) (*framework.CRUDResource, error) {
	c := framework.CRUDResourceConfig{
		Logger: logger,
		Ops:    ops,
	}

	r, err := framework.NewCRUDResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
