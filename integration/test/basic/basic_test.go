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
	"github.com/giantswarm/chart-operator/integration/chartconfig"
	"github.com/giantswarm/chart-operator/integration/release"
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

	chartValuesConfig := chartconfig.ChartValuesConfig{
		Channel:   "5-5-beta",
		Name:      "tb-chart",
		Namespace: "giantswarm",
		Release:   "tb-release",
		//TODO: fix this static VersionBundleVersion
		VersionBundleVersion: "0.2.0",
	}

	config := chartconfig.Config{
		ChartValuesConfig: chartValuesConfig,
	}

	cc, err := chartconfig.NewChartConfig(config)
	if err != nil {
		t.Fatalf("could not create chartconfig %v", err)
	}

	// Setup
	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("could not create logger %v", err)
	}

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
	chartValues, err := cc.ExecuteChartValuesTemplate()
	if err != nil {
		t.Fatalf("could not template chart values %q %v", chartValues, err)
	}

	err = f.InstallResource(cr, chartValues, ":stable")
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
	err = updateChartOperatorResource(cc, helmClient, cr)
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

func updateChartOperatorResource(cc *chartconfig.ChartConfig, helmClient *helmclient.Client, releaseName string) error {
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
	chartValuesConfig := cc.ChartValuesConfig()
	chartValuesConfig.Channel = "5-6-beta"
	cc.SetChartValuesConfig(chartValuesConfig)
	l.Log(chartValuesConfig)

	chartValues, err := cc.ExecuteChartValuesTemplate()
	if err != nil {
		return microerror.Mask(err)
	}
	l.Log(chartValues)

	helmClient.UpdateReleaseFromTarball(releaseName, tarballPath,
		helm.UpdateValueOverrides([]byte(chartValues)))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
