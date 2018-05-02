// +build k8srequired

package basic

import (
	"log"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/integration/templates"
)

func TestChartLifecycle(t *testing.T) {
	const testRelease = "tb-release"
	const cr = "chart-operator-resource"

	log.Println("creating test chart CR")
	err := f.InstallResource(cr, templates.ChartOperatorResourceValues, ":stable")
	if err != nil {
		t.Fatalf("could not install %q %v", cr, err)
	}

	err = waitForReleaseStatus(helmClient, testRelease, "DEPLOYED")
	if err != nil {
		t.Fatalf("could not get release status of %q %v", testRelease, err)
	}
	log.Println("test chart succesfully deployed")

	err = helmClient.DeleteRelease(cr)
	if err != nil {
		t.Fatalf("could not delete %q %v", cr, err)
	}

	err = waitForReleaseStatus(helmClient, testRelease, "DELETED")
	if err != nil {
		t.Fatalf("could not get release status of %q %v", testRelease, err)
	}
	log.Println("test chart succesfully deleted")
}

func waitForReleaseStatus(helmClient *helmclient.Client, release string, status string) error {
	operation := func() error {
		rc, err := helmClient.GetReleaseContent(release)
		if err != nil {
			return microerror.Maskf(err, "could not retrieve release content")
		}
		if rc.Status != status {
			return microerror.Newf("waiting for %q, current %q", status, rc.Status)
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("getting release status %s: %v", t, err)
	}

	return backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
}
