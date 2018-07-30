// +build k8srequired

package basic

import (
	"fmt"
	"testing"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/chart-operator/integration/chart"
	"github.com/giantswarm/chart-operator/integration/release"
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

	gsHelmClient, err := createGsHelmClient()
	if err != nil {
		t.Fatalf("could not create giantswarm helmClient %v", err)
	}

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

	err = release.WaitForStatus(gsHelmClient, testRelease, "DEPLOYED")
	if err != nil {
		t.Fatalf("could not get release status of %q %v", testRelease, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deployed", testRelease))

	// Test Update
	l.Log("level", "debug", "message", fmt.Sprintf("updating %s", cr))
	err = r.UpdateResource(cr, templates.UpdatedChartOperatorResourceValues, "stable")
	if err != nil {
		t.Fatalf("could not update %q %v", cr, err)
	}

	err = release.WaitForVersion(gsHelmClient, testRelease, "5.6.0")
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

	err = release.WaitForStatus(gsHelmClient, testRelease, "DELETED")
	if !helmclient.IsReleaseNotFound(err) {
		t.Fatalf("%q not succesfully deleted %v", testRelease, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deleted", testRelease))
}

func createGsHelmClient() (*helmclient.Client, error) {
	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		return nil, microerror.Maskf(err, "could not create logger")
	}

	c := helmclient.Config{
		Logger:          l,
		K8sClient:       f.K8sClient(),
		RestConfig:      f.RestConfig(),
		TillerNamespace: "giantswarm",
	}

	gsHelmClient, err := helmclient.New(c)
	if err != nil {
		return nil, microerror.Maskf(err, "could not create helmClient")
	}

	return gsHelmClient, nil
}
