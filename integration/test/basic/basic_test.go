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
	const release = "tb-release"

	log.Println("creating test chart CR")
	err := f.InstallResource("chart-operator-resource", templates.ChartOperatorResourceValues, ":stable")
	if err != nil {
		t.Fatalf("could not install chart-operator-resource-chart %v", err)
	}

	err = waitForReleaseStatus(helmClient, release, "DEPLOYED")
	if err != nil {
		t.Fatal("could not get release status", err)
	}
	log.Println("test chart succesfully deployed")

	log.Println("deleting test chart CR")
	err = helmClient.DeleteRelease(release)
	if err != nil {
		t.Fatalf("could not delete chart-operator-resource-chart %v", err)
	}

	err = waitForReleaseStatus(helmClient, release, "DELETED")
	if err != nil {
		t.Fatal("could not get release status", err)
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
