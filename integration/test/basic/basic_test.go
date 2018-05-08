// +build k8srequired

package basic

import (
	"log"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/chart-operator/integration/templates"
)

func TestChartLifecycle(t *testing.T) {
	const testRelease = "tb-release"
	const cr = "chart-operator-resource"

	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("could not create logger %v", err)
	}

	c := helmclient.Config{
		Logger:          l,
		K8sClient:       f.K8sClient(),
		RestConfig:      f.RestConfig(),
		TillerNamespace: "giantswarm",
	}
	gsHelmClient, err := helmclient.New(c)
	if err != nil {
		t.Fatalf("could not create helmClient %v", err)
	}

	log.Printf("creating %q", cr)
	err = f.InstallResource(cr, templates.ChartOperatorResourceValues, ":stable")
	if err != nil {
		t.Fatalf("could not install %q %v", cr, err)
	}

	err = waitForReleaseStatus(gsHelmClient, testRelease, "DEPLOYED")
	if err != nil {
		t.Fatalf("could not get release status of %q %v", testRelease, err)
	}
	log.Printf("%q succesfully deployed", testRelease)

	log.Printf("deleting %q", cr)
	err = helmClient.DeleteRelease(cr)
	if err != nil {
		t.Fatalf("could not delete %q %v", cr, err)
	}

	err = waitForReleaseStatus(gsHelmClient, testRelease, "DELETED")
	if !helmclient.IsReleaseNotFound(err) {
		t.Fatalf("%q not succesfully deleted %v", testRelease, err)
	}
	log.Printf("%q succesfully deleted", testRelease)
}

func waitForReleaseStatus(gsHelmClient *helmclient.Client, release string, status string) error {
	operation := func() error {
		rc, err := gsHelmClient.GetReleaseContent(release)
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

	b := framework.NewExponentialBackoff(framework.ShortMaxWait, framework.LongMaxInterval)
	return backoff.RetryNotify(operation, b, notify)
}
