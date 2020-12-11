package collector

import (
	"github.com/giantswarm/exporterkit/collector"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type SetConfig struct {
	K8sClient  k8sclient.Interface
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

	var helmV2ReleaseCollector *HelmV2Release
	{
		c := HelmV2ReleaseConfig{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			TillerNamespace: config.TillerNamespace,
		}

		helmV2ReleaseCollector, err = NewHelmV2Release(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var orphanConfigMapCollector *OrphanConfigMap
	{
		c := OrphanConfigMapConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		orphanConfigMapCollector, err = NewOrphanConfigMap(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var orphanSecretCollector *OrphanSecret
	{
		c := OrphanSecretConfig{
			G8sClient: config.K8sClient.G8sClient(),
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,
		}

		orphanSecretCollector, err = NewOrphanSecret(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var collectorSet *collector.Set
	{
		c := collector.SetConfig{
			Collectors: []collector.Interface{
				helmV2ReleaseCollector,
				orphanConfigMapCollector,
				orphanSecretCollector,
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
