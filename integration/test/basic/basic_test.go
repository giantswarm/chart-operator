// +build k8srequired

package basic

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/spf13/afero"
	"k8s.io/helm/pkg/helm"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/chart-operator/integration/templates"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

func TestChartLifecycle(t *testing.T) {
	const testRelease = "tb-release"
	const cr = "chart-operator-resource"

	// Setup helm client for giantswarm tiller
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

	// Test Creation
	l.Log("level", "debug", "message", fmt.Sprintf("creating %s", cr))
	err = f.InstallResource(cr, templates.ChartOperatorResourceValues, ":stable")
	if err != nil {
		t.Fatalf("could not install %q %v", cr, err)
	}

	err = waitForReleaseStatus(gsHelmClient, testRelease, "DEPLOYED")
	if err != nil {
		t.Fatalf("could not get release status of %q %v", testRelease, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deployed", testRelease))

	// Test Update
	log.Printf("updating %q", cr)
	err = updateChartOperatorResource(helmClient, cr)
	if err != nil {
		t.Fatalf("could not update %q %v", cr, err)
	}

	err = waitForReleaseStatus(gsHelmClient, testRelease, "PENDING_UPGRADE")
	if err != nil {
		t.Fatalf("could not get release status of %q %v", testRelease, err)
	}
	err = waitForReleaseStatus(gsHelmClient, testRelease, "DEPLOYED")
	if err != nil {
		t.Fatalf("could not get release status of %q %v", testRelease, err)
	}
	log.Println("test chart succesfully updated")

	// Test Deletion
	l.Log("level", "debug", "message", fmt.Sprintf("deleting %s", cr))
	err = helmClient.DeleteRelease(cr)
	if err != nil {
		t.Fatalf("could not delete %q %v", cr, err)
	}

	err = waitForReleaseStatus(gsHelmClient, testRelease, "DELETED")
	if !helmclient.IsReleaseNotFound(err) {
		t.Fatalf("%q not succesfully deleted %v", testRelease, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deleted", testRelease))
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

func updateChartOperatorResource(helmClient *helmclient.Client, releaseName string) error {
	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		return microerror.Mask(err)
	}

	c := apprclient.Config{
		Fs:     afero.NewOsFs(),
		Logger: l,

		Address:      "https://quay.io",
		Organization: "giantswarm",
	}

	a, err := apprclient.New(c)
	if err != nil {
		return microerror.Mask(err)
	}

	tarballPath, err := a.PullChartTarball(fmt.Sprintf("%s-chart", releaseName), "stable")
	if err != nil {
		return microerror.Mask(err)
	}

	helmClient.UpdateReleaseFromTarball(releaseName, tarballPath,
		helm.UpdateValueOverrides([]byte(templates.UpdatedChartOperatorResourceValues)),
		helm.UpgradeWait(true))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
