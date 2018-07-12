// +build k8srequired

package release

import (
	"log"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
)

func WaitForStatus(gsHelmClient *helmclient.Client, release string, status string) error {
	operation := func() error {
		rc, err := gsHelmClient.GetReleaseContent(release)
		if err != nil {
			return microerror.Mask(err)
		}
		if rc.Status != status {
			return microerror.Maskf(releaseStatusNotMatchingError, "waiting for %q, current %q", status, rc.Status)
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("getting release status %s: %v", t, err)
	}

	b := framework.NewExponentialBackoff(framework.ShortMaxWait, framework.LongMaxInterval)
	err := backoff.RetryNotify(operation, b, notify)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func WaitForVersion(gsHelmClient *helmclient.Client, release string, version string) error {
	operation := func() error {
		rh, err := gsHelmClient.GetReleaseHistory(release)
		if err != nil {
			return microerror.Mask(err)
		}
		if rh.Version != version {
			return microerror.Maskf(releaseVersionNotMatchingError, "waiting for %q, current %q", version, rh.Version)
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("getting release version %s: %v", t, err)
	}

	b := framework.NewExponentialBackoff(framework.ShortMaxWait, framework.LongMaxInterval)
	err := backoff.RetryNotify(operation, b, notify)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}
