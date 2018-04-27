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

func TestChartInstalled(t *testing.T) {
	err := f.InstallResource("chart-operator-resource", templates.ChartOperatorResourceValues, ":stable")
	if err != nil {
		t.Fatalf("could not install chart-operator-resource-chart %v", err)
	}

	var rc *helmclient.ReleaseContent
	operation := func() error {
		rc, err = helmClient.GetReleaseContent("tb-release")
		if err != nil {
			return microerror.Maskf(err, "could not retrieve release content")
		}
		if rc.Status == "PENDING_INSTALL" {
			return microerror.Newf("release still not installed")
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("waiting for release %s: %v", t, err)
	}

	err = backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
	if err != nil {
		t.Fatal("expected nil found", err)
	}

	expectedStatus := "DEPLOYED"
	if rc.Status != expectedStatus {
		t.Fatalf("unexpected chart status, want %q, got %q", expectedStatus, rc.Status)
	}
}
