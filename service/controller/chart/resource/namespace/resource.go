package namespace

import (
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

const (
	Name = "namespace"

	// defaultK8sWaitTimeout is how long to wait for the Kubernetes API when
	// installing or updating a release before moving to process the next CR.
	defaultK8sWaitTimeout = 10 * time.Second
)

type Config struct {
	// Dependencies.
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	// Settings.
	K8sWaitTimeout time.Duration
}

type Resource struct {
	// Dependencies.
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	// Settings.
	k8sWaitTimeout time.Duration
}

// New creates a new configured tcnamespace resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// Settings.
	if config.K8sWaitTimeout == 0 {
		config.K8sWaitTimeout = defaultK8sWaitTimeout
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		k8sWaitTimeout: config.K8sWaitTimeout,
	}

	return r, nil
}

func (r Resource) Name() string {
	return Name
}
