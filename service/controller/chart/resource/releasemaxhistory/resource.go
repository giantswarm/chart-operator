package releasemaxhistory

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/v2/service/internal/clientpair"
)

const (
	Name = "releasemaxhistory"
)

type Config struct {
	// Dependencies.
	ClientPair *clientpair.ClientPair
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
}

type Resource struct {
	// Dependencies.
	clientPair *clientpair.ClientPair
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
}

// New creates a new configured releasemaxhistory resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.ClientPair == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClientPair must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		clientPair: config.ClientPair,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r Resource) Name() string {
	return Name
}
