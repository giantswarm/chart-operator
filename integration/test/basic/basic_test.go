// +build k8srequired

package basic

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/chart-operator/integration/chart"
	"github.com/giantswarm/chart-operator/integration/templates"
)

func TestChartLifecycle(t *testing.T) {
	const testRelease = "tb-release"
	const cr = "apiextensions-chart-config-e2e"

	charts := []chart.Chart{
		{
			Channel: "5-5-beta",
			Release: "5.5.5",
			Tarball: "/e2e/fixtures/tb-chart-5.5.5.tgz",
			Name:    "tb-chart",
		},
		{
			Channel: "5-6-beta",
			Release: "5.6.0",
			Tarball: "/e2e/fixtures/tb-chart-5.6.0.tgz",
			Name:    "tb-chart",
		},
	}

	// Setup
	err = chart.Push(f, charts)
	if err != nil {
		t.Fatalf("could not push inital charts to cnr %v", err)
	}

	// Test Creation
	l.Log("level", "debug", "message", fmt.Sprintf("creating %s", cr))
	err = r.InstallResource(cr, templates.ChartOperatorResourceValues, "stable")
	if err != nil {
		t.Fatalf("could not install %q %v", cr, err)
	}

	err = r.WaitForStatus(testRelease, "DEPLOYED")
	if err != nil {
		t.Fatalf("could not get release status of %q %v", testRelease, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deployed", testRelease))

	// Test Update
	l.Log("level", "debug", "message", fmt.Sprintf("updating %s", cr))
	err = updateChartOperatorResource(helmClient, cr)
	if err != nil {
		t.Fatalf("could not update %q %v", cr, err)
	}

	err = r.WaitForVersion(testRelease, "5.6.0")
	if err != nil {
		t.Fatalf("could not get release version of %q %v", testRelease, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully updated", testRelease))

	// Test Deletion
	l.Log("level", "debug", "message", fmt.Sprintf("deleting %s", cr))
	err = helmClient.DeleteRelease(cr)
	if err != nil {
		t.Fatalf("could not delete %q %v", cr, err)
	}

	err = r.WaitForStatus(testRelease, "DELETED")
	if !helmclient.IsReleaseNotFound(err) {
		t.Fatalf("%q not succesfully deleted %v", testRelease, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deleted", testRelease))
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
		helm.UpdateValueOverrides([]byte(templates.UpdatedChartOperatorResourceValues)))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
