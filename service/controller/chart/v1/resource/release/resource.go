package release

import (
	"context"
	"reflect"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/controller"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
)

const (
	// Name is the identifier of the resource.
	Name = "releasev1"

	// helmDeployedStatus is the deployed status for Helm Releases.
	helmDeployedStatus = "DEPLOYED"
)

// Config represents the configuration used to create a new release resource.
type Config struct {
	// Dependencies.
	Fs         afero.Fs
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
}

// Resource implements the chart resource.
type Resource struct {
	// Dependencies.
	fs         afero.Fs
	helmClient helmclient.Interface
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*controller.Patch, error) {
	return nil, nil
}

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteChange interface{}) error {
	return nil
}

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateChange interface{}) error {
	return nil
}

// New creates a new configured chart resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		fs:         config.Fs,
		helmClient: config.HelmClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

// Equals asseses the equality of ReleaseStates with regards to distinguishing fields.
func (a *ReleaseState) Equals(b ReleaseState) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Status != b.Status {
		return false
	}
	if !reflect.DeepEqual(a.Values, b.Values) {
		return false
	}
	if a.Version != b.Version {
		return false
	}
	return true
}

// IsEmpty checks if a ReleaseState is empty.
func (c *ReleaseState) IsEmpty() bool {
	return c.Equals(ReleaseState{})
}

func toReleaseState(v interface{}) (ReleaseState, error) {
	if v == nil {
		return ReleaseState{}, nil
	}

	releaseState, ok := v.(*ReleaseState)
	if !ok {
		return ReleaseState{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", releaseState, v)
	}

	return *releaseState, nil
}
