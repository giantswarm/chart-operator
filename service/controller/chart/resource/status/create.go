package status

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/to"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/chart-operator/v2/pkg/annotation"
	"github.com/giantswarm/chart-operator/v2/pkg/project"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/controllercontext"
	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
)

func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCustomResource(obj)
	if err != nil {
		return microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	releaseName := key.ReleaseName(cr)
	r.logger.Debugf(ctx, "getting status for release %#q", releaseName)

	releaseContent, err := r.helmClient.GetReleaseContent(ctx, key.Namespace(cr), releaseName)
	if helmclient.IsReleaseNotFound(err) {
		r.logger.Debugf(ctx, "release %#q not found", releaseName)

		// If something goes wrong outside of Helm we add that to the
		// controller context in the release resource. So we include this
		// information in the CR status.
		if cc.Status.Reason != "" {
			status := v1alpha1.ChartStatus{
				Reason: cc.Status.Reason,
				Release: v1alpha1.ChartStatusRelease{
					Status: cc.Status.Release.Status,
				},
			}

			err = r.setStatus(ctx, cr, status)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}

		// There is no Helm release and no information added to the
		// context by the release resource.
		//
		// Return early. We will retry on the next execution.
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	var status, reason string
	{
		if key.IsCordoned(cr) {
			status = releaseStatusCordoned
			reason = key.CordonReason(cr)
		} else {
			status = releaseContent.Status
			if releaseContent.Status != helmclient.StatusDeployed {
				if cc.Status.Release.FailedMaxAttempts {
					reason = fmt.Sprintf("Release has failed %d times.\nReason: %s",
						project.ReleaseFailedMaxAttempts,
						releaseContent.Description)
				} else {
					reason = releaseContent.Description
				}
			}
		}
	}

	desiredStatus := v1alpha1.ChartStatus{
		AppVersion: releaseContent.AppVersion,
		Reason:     reason,
		Release: v1alpha1.ChartStatusRelease{
			Revision: to.IntP(releaseContent.Revision),
			Status:   status,
		},
		Version: releaseContent.Version,
	}
	if !releaseContent.LastDeployed.IsZero() {
		// We convert the timestamp to the nearest second to match the value in
		// the chart CR status.
		lastDeployed := releaseContent.LastDeployed.Unix()
		desiredStatus.Release.LastDeployed = &metav1.Time{Time: time.Unix(lastDeployed, 0)}
	}

	if !equals(desiredStatus, key.ChartStatus(cr)) {
		err = r.setStatus(ctx, cr, desiredStatus)
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		r.logger.Debugf(ctx, "status for release %#q already set to %#q", releaseName, releaseContent.Status)
	}

	return nil
}

func (r *Resource) getAuthToken(ctx context.Context) (string, error) {
	secret, err := r.k8sClient.CoreV1().Secrets(namespace).Get(ctx, authTokenName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		// There is no auth token secret. It may not have been created yet. Or the app CR is using InCluster.
		r.logger.Debugf(ctx, "no auth token secret found skip calling webhook")
		return "", nil
	} else if err != nil {
		return "", microerror.Mask(err)
	}

	return string(secret.Data[token]), nil
}

func (r *Resource) setStatus(ctx context.Context, cr v1alpha1.Chart, status v1alpha1.ChartStatus) error {
	if url, ok := cr.GetAnnotations()[annotation.Webhook]; ok {
		authToken, err := r.getAuthToken(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		err = updateAppStatus(url, authToken, status, r.httpClientTimeout)
		if err != nil {
			r.logger.Errorf(ctx, err, "sending webhook to %#q failed", url)
		}
	}

	r.logger.Debugf(ctx, "setting status for release %#q status to %#q", key.ReleaseName(cr), status.Release.Status)

	// Get chart CR again to ensure the resource version is correct.
	var currentCR v1alpha1.Chart

	err := r.ctrlClient.Get(
		ctx,
		types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace},
		&currentCR,
	)
	if err != nil {
		return microerror.Mask(err)
	}

	currentCR.Status = status

	err = r.ctrlClient.Status().Update(ctx, &currentCR)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "set status for release %#q", key.ReleaseName(cr))

	return nil
}

func updateAppStatus(webhookURL, authToken string, status v1alpha1.ChartStatus, timeout time.Duration) error {
	request := Request{
		AppVersion: status.AppVersion,
		Reason:     status.Reason,
		Status:     status.Release.Status,
		Version:    status.Version,
	}
	if status.Release.LastDeployed != nil {
		request.LastDeployed = *status.Release.LastDeployed
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return microerror.Mask(err)
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodPatch, webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		return microerror.Mask(err)
	}

	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return microerror.Mask(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return microerror.Maskf(wrongStatusError, "expected http status '%d', got '%d'", http.StatusOK, resp.StatusCode)
	}

	return nil
}
