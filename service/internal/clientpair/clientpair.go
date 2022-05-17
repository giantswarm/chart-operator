package clientpair

import (
	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
)

const (
	// privateNamespace defines GS-protected namespace the prvHelmClient
	// is meant for. For App CRs created outside this namespace, the
	// pubHelmClient should be used.
	privateNamespace = "giantswarm"
)

type ClientPairConfig struct {
	PrvHelmClient helmclient.Interface
	PubHelmClient helmclient.Interface
}

type ClientPair struct {
	prvHelmClient helmclient.Interface
	pubHelmClient helmclient.Interface
}

func NewClientPair(config ClientPairConfig) (*ClientPair, error) {
	if config.PrvHelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.PrvHelmClient must not be empty", config)
	}
	if config.PubHelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.PubHelmClient must not be empty", config)
	}

	cp := &ClientPair{
		prvHelmClient: config.PrvHelmClient,
		pubHelmClient: config.PubHelmClient,
	}

	return cp, nil
}

// Get determines which client to use base on the namespace the corresponding App CR
// is located in. For Workload Cluster private and public Helm clients are identical
func (cp *ClientPair) Get(cr v1alpha1.Chart) helmclient.Interface {
	if key.AppNamespace(cr) == privateNamespace {
		return cp.prvHelmClient
	}

	return cp.pubHelmClient
}
