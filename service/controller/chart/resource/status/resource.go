package status

import (
	"time"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/to"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/chart-operator/v2/service/internal/clientpair"
)

const (
	Name = "status"

	authTokenName = "auth-token"
	// defaultHTTPClientTimeout is the timeout when updating app status.
	defaultHTTPClientTimeout = 5
	namespace                = "giantswarm"
	releaseStatusCordoned    = "CORDONED"
	token                    = "token"
)

// Config represents the configuration used to create a new status resource.
type Config struct {
	CtrlClient  client.Client
	HelmClients *clientpair.ClientPair
	K8sClient   kubernetes.Interface
	Logger      micrologger.Logger

	HTTPClientTimeout time.Duration
}

// Resource implements the status resource.
type Resource struct {
	ctrlClient  client.Client
	helmClients *clientpair.ClientPair
	k8sClient   kubernetes.Interface
	logger      micrologger.Logger

	httpClientTimeout time.Duration
}

// New creates a new configured status resource.
func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.HelmClients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.HTTPClientTimeout == 0 {
		config.HTTPClientTimeout = defaultHTTPClientTimeout
	}

	r := &Resource{
		ctrlClient:  config.CtrlClient,
		helmClients: config.HelmClients,
		k8sClient:   config.K8sClient,
		logger:      config.Logger,

		httpClientTimeout: config.HTTPClientTimeout,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

// equals asseses the equality of ChartStatuses with regards to distinguishing
// fields.
func equals(a, b v1alpha1.ChartStatus) bool {
	if a.AppVersion != b.AppVersion {
		return false
	}

	var lastDeployedA, lastDeployedB int64

	if a.Release.LastDeployed != nil {
		lastDeployedA = a.Release.LastDeployed.Unix()
	}
	if b.Release.LastDeployed != nil {
		lastDeployedB = b.Release.LastDeployed.Unix()
	}
	if lastDeployedA != lastDeployedB {
		return false
	}

	if a.Reason != b.Reason {
		return false
	}

	var revisionA, revisionB int

	if a.Release.Revision != nil {
		revisionA = to.Int(a.Release.Revision)
	}
	if b.Release.Revision != nil {
		revisionB = to.Int(b.Release.Revision)
	}
	if revisionA != revisionB {
		return false
	}

	if a.Release.Status != b.Release.Status {
		return false
	}
	if a.Version != b.Version {
		return false
	}

	return true
}
