package chart

import (
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v7/pkg/resource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/crud"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v7/pkg/resource/wrapper/retryresource"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/chart-operator/v2/service/controller/chart/resource/namespace"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/resource/release"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/resource/releasemaxhistory"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/resource/status"

	"github.com/giantswarm/chart-operator/v2/service/internal/clientpair"
)

type chartResourcesConfig struct {
	// Dependencies.
	Fs          afero.Fs
	CtrlClient  client.Client
	HelmClients *clientpair.ClientPair
	K8sClient   kubernetes.Interface
	Logger      micrologger.Logger

	// Settings.
	HTTPClientTimeout time.Duration
	K8sWaitTimeout    time.Duration
	MaxRollback       int
	TillerNamespace   string
}

func newChartResources(config chartResourcesConfig) ([]resource.Interface, error) {
	var err error

	// Dependencies.
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.HelmClients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClients must not be empty", config)
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

	var namespaceResource resource.Interface
	{
		c := namespace.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			K8sWaitTimeout: config.K8sWaitTimeout,
		}

		namespaceResource, err = namespace.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var releaseResource resource.Interface
	{
		c := release.Config{
			// Dependencies
			Fs:          config.Fs,
			CtrlClient:  config.CtrlClient,
			HelmClients: config.HelmClients,
			K8sClient:   config.K8sClient,
			Logger:      config.Logger,

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

	var releaseMaxHistoryResource resource.Interface
	{
		c := releasemaxhistory.Config{
			// Dependencies
			HelmClients: config.HelmClients,
			K8sClient:   config.K8sClient,
			Logger:      config.Logger,
		}

		releaseMaxHistoryResource, err = releasemaxhistory.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var statusResource resource.Interface
	{
		c := status.Config{
			CtrlClient:  config.CtrlClient,
			HelmClients: config.HelmClients,
			K8sClient:   config.K8sClient,
			Logger:      config.Logger,

			HTTPClientTimeout: config.HTTPClientTimeout,
		}

		statusResource, err = status.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// namespace creates the release namespace and allows setting metadata.
		namespaceResource,
		// release max history ensures not too many helm release secrets are created.
		releaseMaxHistoryResource,
		// release manages Helm releases and is the most important resource.
		releaseResource,
		// status resource manages the chart CR status.
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
