package collector

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

type SetConfig struct {
	G8sClient  versioned.Interface
	K8sClient  kubernetes.Interface
	HelmClient *helmclient.Client
	Logger     micrologger.Logger

	TillerNamespace string
}

// Set is basically only a wrapper for the operator's collector implementations.
// It eases the iniitialization and prevents some weird import mess so we do not
// have to alias packages.
type Set struct {
	*collector.Set
}

func NewSet(config SetConfig) (*Set, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.TillerNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TillerNamespace must not be empty", config)
	}

	var err error

	var chartResourceCollector *ChartResource
	{
		c := ChartResourceConfig{
			G8sClient: config.G8sClient,
			Logger:    config.Logger,
		}

		chartResourceCollector, err = NewChartResource(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tillerMaxHistoryCollector *TillerMaxHistory
	{
		c := TillerMaxHistoryConfig{
			G8sClient: config.G8sClient,
			K8sClient: config.K8sClient,
			Logger:    config.Logger,

			TillerNamespace: config.TillerNamespace,
		}

		tillerMaxHistoryCollector, err = NewTillerMaxHistory(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tillerReachableCollector *TillerReachable
	{
		c := TillerReachableConfig{
			G8sClient:  config.G8sClient,
			HelmClient: config.HelmClient,
			Logger:     config.Logger,

			TillerNamespace: config.TillerNamespace,
		}

		tillerReachableCollector, err = NewTillerReachable(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: []collector.Interface{
				chartResourceCollector,
				tillerMaxHistoryCollector,
				tillerReachableCollector,
			},
			Logger: config.Logger,
		}

		collectorSet, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Set{
		Set: collectorSet,
	}

	return s, nil
}
