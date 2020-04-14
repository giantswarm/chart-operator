package release

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/service/controller/chart/key"
)

const (
	// Name is the identifier of the resource.
	Name = "release"

	// helmDeployedStatus is the deployed status for Helm Releases.
	helmDeployedStatus = "DEPLOYED"
	// helmFailedStatus is the failed status for Helm Releases.
	helmFailedStatus = "FAILED"
	// releaseNotInstalledStatus is set in the CR status when there is no Helm
	// Release to check.
	releaseNotInstalledStatus = "Not installed"
)

var (
	// releaseTransitionStatuses is used to determine if the Helm Release is
	// currently being updated.
	releaseTransitionStatuses = map[string]bool{
		"DELETING":         true,
		"PENDING_INSTALL":  true,
		"PENDING_UPGRADE":  true,
		"PENDING_ROLLBACK": true,
	}
)

// Config represents the configuration used to create a new release resource.
type Config struct {
	// Dependencies.
	Fs         afero.Fs
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger
}

// Resource implements the chart resource.
type Resource struct {
	// Dependencies.
	fs         afero.Fs
	g8sClient  versioned.Interface
	helmClient helmclient.Interface
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger
}

// New creates a new configured chart resource.
func New(config Config) (*Resource, error) {
	// Dependencies.
	if config.Fs == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Fs must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
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
		g8sClient:  config.G8sClient,
		helmClient: config.HelmClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

// patchAnnotations updates the chart CR annotations if they have changed.
// A patch operation is used because app-operator also sets annotations for
// chart CRs.
func (r *Resource) patchAnnotations(ctx context.Context, cr v1alpha1.Chart, releaseState ReleaseState) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("patching annotations for chart CR %#q in namespace %#q", cr.Name, cr.Namespace))

	// Get chart CR again to ensure the resource version and annotations
	// are correct.
	currentCR, err := r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).Get(cr.Name, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	currentChecksum := key.ValuesMD5ChecksumAnnotation(*currentCR)

	if releaseState.ValuesMD5Checksum != currentChecksum {
		patches := []Patch{}

		if len(currentCR.Annotations) == 0 {
			patches = append(patches, Patch{
				Op:    "add",
				Path:  "/metadata/annotations",
				Value: map[string]string{},
			})
		}

		patches = append(patches, Patch{
			Op:    "add",
			Path:  fmt.Sprintf("/metadata/annotations/%s", replaceToEscape(key.ValuesMD5ChecksumAnnotationName)),
			Value: releaseState.ValuesMD5Checksum,
		})

		bytes, err := json.Marshal(patches)
		if err != nil {
			return microerror.Mask(err)
		}

		_, err = r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).Patch(cr.Name, types.JSONPatchType, bytes)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("patched annotations for chart CR %#q in namespace %#q", cr.Name, cr.Namespace))

	} else {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("no need to patch annotations for chart CR %#q in namespace %#q", cr.Name, cr.Namespace))
	}

	return nil
}

// addStatusToContext adds the status to the controller context. It will be
// used to set the CR status in the status resource.
func addStatusToContext(cc *controllercontext.Context, reason, status string) {
	cc.Status = controllercontext.Status{
		Reason: reason,
		Release: controllercontext.Release{
			Status: status,
		},
	}
}

// equals asseses the equality of ReleaseStates with regards to distinguishing fields.
func equals(a, b ReleaseState) bool {
	if a.Name != b.Name {
		return false
	}
	if a.Status != b.Status {
		return false
	}
	if a.ValuesMD5Checksum != b.ValuesMD5Checksum {
		return false
	}
	if a.Version != b.Version {
		return false
	}
	return true
}

// isEmpty checks if a ReleaseState is empty.
func isEmpty(c ReleaseState) bool {
	return equals(c, ReleaseState{})
}

func isReleaseInTransitionState(r ReleaseState) bool {
	return releaseTransitionStatuses[r.Status]
}

func isReleaseModified(a, b ReleaseState) bool {
	if isEmpty(a) {
		return false
	}

	// Values have changed so we need to update the Helm Release.
	if a.ValuesMD5Checksum != b.ValuesMD5Checksum {
		return true
	}

	if a.Status != b.Status {
		return true
	}

	if a.Version != b.Version {
		return true
	}

	return false
}

func replaceToEscape(from string) string {
	return strings.Replace(from, "/", "~1", -1)
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
