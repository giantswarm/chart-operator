package releasemaxhistory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/v2/pkg/project"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
)

// EnsureCreated checks if the helm release has failed the max number of
// attempts. If so we delete the oldest revision if it is over 1 minute old.
// So we still retry the update but at a reduced rate. This is needed because
// the max history setting for Helm update does not count failures.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "finding out if release %#q in namespace %#q has failed max attempts", key.ReleaseName(cr), key.Namespace(cr))

	history, err := r.getReleaseHistory(ctx, key.Namespace(cr), key.ReleaseName(cr))
	if err != nil {
		return microerror.Mask(err)
	}

	failedMaxAttempts, err := isReleaseFailedMaxAttempts(ctx, history)
	if err != nil {
		return microerror.Mask(err)
	}

	cc.Status.Release.FailedMaxAttempts = failedMaxAttempts
	if !failedMaxAttempts {
		r.logger.Debugf(ctx, "release %#q has not failed max attempts", key.ReleaseName(cr))
		return nil
	}

	r.logger.Debugf(ctx, "release %#q has failed max attempts", key.ReleaseName(cr))

	secretDeleted, err := r.deleteFailedRelease(ctx, key.Namespace(cr), key.ReleaseName(cr), history)
	if err != nil {
		return microerror.Mask(err)
	}
	if secretDeleted {
		// We deleted a failed release secret. So we can try to update the release again.
		cc.Status.Release.FailedMaxAttempts = false
	}

	return nil
}

func (r *Resource) deleteFailedRelease(ctx context.Context, namespace, releaseName string, history []helmclient.ReleaseHistory) (bool, error) {
	rev := history[project.ReleaseFailedMaxAttempts-1]

	r.logger.Debugf(ctx, "deleting failed revision %d for release %#q", rev.Revision, releaseName)

	selectors := []string{
		"owner=helm",
		"status=failed",
		fmt.Sprintf("%s=%s", "name", releaseName),
		fmt.Sprintf("%s=%d", "version", rev.Revision),
	}
	lo := metav1.ListOptions{
		LabelSelector: strings.Join(selectors, ","),
	}
	secrets, err := r.k8sClient.CoreV1().Secrets(namespace).List(ctx, lo)
	if err != nil {
		return false, microerror.Mask(err)
	}
	if len(secrets.Items) != 1 {
		return false, microerror.Maskf(executionFailedError, "expected 1 release secret got %d", len(secrets.Items))
	}

	secret := secrets.Items[0]
	diff := time.Since(secret.CreationTimestamp.Time)
	if diff.Minutes() < 1 {
		r.logger.Debugf(ctx, "revision %d for release %#q is < 1 minutes old", rev.Revision, releaseName)
		r.logger.Debugf(ctx, "canceling resource")
		return false, nil
	}

	err = r.k8sClient.CoreV1().Secrets(secret.Namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.Debugf(ctx, "already deleted revision %d for release %#q", rev.Revision, releaseName)
		return true, nil
	} else if err != nil {
		return false, microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "deleted failed revision %d for release %#q", rev.Revision, releaseName)

	return true, nil
}

func (r *Resource) getReleaseHistory(ctx context.Context, namespace, releaseName string) ([]helmclient.ReleaseHistory, error) {
	history, err := r.helmClient.GetReleaseHistory(ctx, namespace, releaseName)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Sort history by descending revision number.
	sort.Slice(history, func(i, j int) bool {
		return history[i].Revision > history[j].Revision
	})

	return history, nil
}

func isReleaseFailedMaxAttempts(ctx context.Context, history []helmclient.ReleaseHistory) (bool, error) {
	if len(history) < project.ReleaseFailedMaxAttempts {
		return false, nil
	}

	for i := 0; i < project.ReleaseFailedMaxAttempts; i++ {
		if history[i].Status != helmclient.StatusFailed {
			return false, nil
		}
	}

	// All failed so we exceeded the max attempts.
	return true, nil
}
