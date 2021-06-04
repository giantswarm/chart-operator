package release

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/v2/pkg/annotation"
	"github.com/giantswarm/chart-operator/v2/pkg/project"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
)

const (
	// Name is the identifier of the resource.
	Name = "release"

	// defaultK8sWaitTimeout is how long to wait for the Kubernetes API when
	// installing or updating a release before moving to process the next CR.
	defaultK8sWaitTimeout = 10 * time.Second

	// alreadyExistsStatus is set in the CR status when it failed to create
	// a manifest object because it exists already.
	alreadyExistsStatus = "already-exists"

	// invalidManifestStatus is set in the CR status when it failed to create
	// manifest objects with helm resources.
	invalidManifestStatus = "invalid-manifest"

	// releaseNotInstalledStatus is set in the CR status when there is no Helm
	// Release to check.
	releaseNotInstalledStatus = "not-installed"

	// unknownError when a release fails for unknown reasons.
	unknownError = "unknown-error"

	// validationFailedStatus is set in the CR status when it failed to pass
	// OpenAPI validation on release manifest.
	validationFailedStatus = "validation-failed"
)

// Config represents the configuration used to create a new release resource.
type Config struct {
	// Dependencies.
	Fs         afero.Fs
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	// Settings.
	K8sWaitTimeout  time.Duration
	MaxRollback     int
	TillerNamespace string
}

// Resource implements the chart resource.
type Resource struct {
	// Dependencies.
	fs         afero.Fs
	g8sClient  versioned.Interface
	helmClient helmclient.Interface
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger

	// Settings.
	k8sWaitTimeout  time.Duration
	maxRollback     int
	tillerNamespace string
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

	// Settings.
	if config.K8sWaitTimeout == 0 {
		config.K8sWaitTimeout = defaultK8sWaitTimeout
	}
	if config.TillerNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TillerNamespace must not be empty", config)
	}

	r := &Resource{
		// Dependencies.
		fs:         config.Fs,
		g8sClient:  config.G8sClient,
		helmClient: config.HelmClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,

		// Settings.
		k8sWaitTimeout:  config.K8sWaitTimeout,
		maxRollback:     config.MaxRollback,
		tillerNamespace: config.TillerNamespace,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) findHelmV2ConfigMaps(ctx context.Context, releaseName string) (bool, error) {
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", "NAME", releaseName, "OWNER", "TILLER"),
	}

	// Check whether there are still helm2 release configmaps.
	cms, err := r.k8sClient.CoreV1().ConfigMaps(r.tillerNamespace).List(ctx, lo)
	if err != nil {
		return false, microerror.Mask(err)
	}

	return len(cms.Items) > 0, nil
}

func (r *Resource) addAnnotation(ctx context.Context, cr *v1alpha1.Chart, key, value string) error {
	patches := []Patch{}

	if len(cr.Annotations) == 0 {
		patches = append(patches, Patch{
			Op:    "add",
			Path:  "/metadata/annotations",
			Value: map[string]string{},
		})
	}

	patches = append(patches, Patch{
		Op:    "add",
		Path:  fmt.Sprintf("/metadata/annotations/%s", replaceToEscape(key)),
		Value: value,
	})

	bytes, err := json.Marshal(patches)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).Patch(ctx, cr.Name, types.JSONPatchType, bytes, metav1.PatchOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) removeAnnotation(ctx context.Context, cr *v1alpha1.Chart, key string) error {
	if _, ok := cr.GetAnnotations()[key]; !ok {
		// no-op
		return nil
	}

	patches := []Patch{
		{
			Op:   "remove",
			Path: fmt.Sprintf("/metadata/annotations/%s", replaceToEscape(key)),
		},
	}

	bytes, err := json.Marshal(patches)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).Patch(ctx, cr.Name, types.JSONPatchType, bytes, metav1.PatchOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// addHashAnnotation updates the chart CR annotations if they have changed.
// A patch operation is used because app-operator also sets annotations for
// chart CRs.
func (r *Resource) addHashAnnotation(ctx context.Context, cr v1alpha1.Chart, releaseState ReleaseState) error {
	r.logger.Debugf(ctx, "patching annotations for chart CR %#q in namespace %#q", cr.Name, cr.Namespace)

	// Get chart CR again to ensure the resource version and annotations
	// are correct.
	currentCR, err := r.g8sClient.ApplicationV1alpha1().Charts(cr.Namespace).Get(ctx, cr.Name, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	currentChecksum := key.ValuesMD5ChecksumAnnotation(*currentCR)

	if releaseState.ValuesMD5Checksum != currentChecksum {
		err := r.addAnnotation(ctx, currentCR, annotation.ValuesMD5Checksum, releaseState.ValuesMD5Checksum)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "patched annotations for chart CR %#q in namespace %#q", cr.Name, cr.Namespace)

	} else {
		r.logger.Debugf(ctx, "no need to patch annotations for chart CR %#q in namespace %#q", cr.Name, cr.Namespace)
	}

	return nil
}

// isReleaseFailedMaxAttempts checks the release history to see if it has
// failed the max number of attempts. If it has we stop updating. This is
// needed as the max history setting for Helm update does not count failures.
func (r *Resource) isReleaseFailedMaxAttempts(ctx context.Context, namespace, releaseName string) (bool, error) {
	history, err := r.helmClient.GetReleaseHistory(ctx, namespace, releaseName)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if len(history) < project.ReleaseFailedMaxAttempts {
		return false, nil
	}

	// Sort history by descending revision number.
	sort.Slice(history, func(i, j int) bool {
		return history[i].Revision > history[j].Revision
	})

	for i := 0; i < project.ReleaseFailedMaxAttempts; i++ {
		if history[i].Status != helmclient.StatusFailed {
			return false, nil
		}
	}

	// All failed so we exceeded the max attempts.
	return true, nil
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

// isReleaseFailed checks if the release is failed. If the values or version
// has changed we return false and will attempt to update the release. As this
// may fix the problem.
func isReleaseFailed(current, desired ReleaseState) bool {
	result := false

	if !isEmpty(current) {
		// Values have changed so we should try to update even if the release
		// is failed.
		if current.ValuesMD5Checksum != desired.ValuesMD5Checksum {
			return false
		}

		// Version has changed so we should try to update even if the release
		// is failed.
		if current.Version != desired.Version {
			return false
		}

		// Release is failed and should not be updated.
		if current.Status == helmclient.StatusFailed {
			result = true
		}
	}

	return result
}

func isReleaseInTransitionState(r ReleaseState) bool {
	return helmclient.ReleaseTransitionStatuses[r.Status]
}

func isReleaseModified(a, b ReleaseState) bool {
	result := false

	if !isEmpty(a) {
		if a.ValuesMD5Checksum != b.ValuesMD5Checksum {
			result = true
		}

		if a.Status != b.Status {
			result = true
		}

		if a.Version != b.Version {
			result = true
		}
	}

	return result
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

func convertFloat(m map[string]interface{}) {
	for k, val := range m {
		switch floatVal := val.(type) {
		case float64:
			converted := int(floatVal)
			if floatVal == float64(converted) {
				m[k] = converted
			}
		case map[string]interface{}:
			convertFloat(val.(map[string]interface{}))
		default:
			// no-op
		}
	}
}
