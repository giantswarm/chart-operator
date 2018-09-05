// +build k8srequired

package chartvalues

import (
	"fmt"
	"testing"

	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/integration/chart"
	"github.com/giantswarm/chart-operator/integration/chartconfig"
	"github.com/giantswarm/chart-operator/integration/env"
)

const (
	cr                   = "apiextensions-chart-config-e2e"
	namespace            = "giantswarm"
	testChartRelease     = "tb-release"
	testConfigMapName    = "values-configmap"
	testConfigMapRelease = "tb-configmap"
)

func TestChartValues(t *testing.T) {
	charts := []chart.Chart{
		{
			Channel: "1-0-beta",
			Release: "1.0.0",
			Tarball: "/e2e/fixtures/tb-chart-1.0.0.tgz",
			Name:    "tb-chart",
		},
	}

	chartConfigValues := e2etemplates.ApiextensionsChartConfigValues{
		Channel: "1-0-beta",
		ConfigMap: e2etemplates.ApiextensionsChartConfigConfigMap{
			Name:            testConfigMapName,
			Namespace:       namespace,
			ResourceVersion: "1",
		},
		Name:                 "tb-chart",
		Namespace:            namespace,
		Release:              testChartRelease,
		VersionBundleVersion: env.VersionBundleVersion(),
	}
	err := chart.Push(l, h, charts)
	if err != nil {
		t.Fatalf("could not push inital charts to cnr %v", err)
	}

	// Test Creation

	// Install Values ConfigMap
	err = helmClient.InstallFromTarball("/e2e/fixtures/tb-configmap-chart-1.0.0.tgz", namespace, helm.ReleaseName(testConfigMapRelease), helm.ValueOverrides([]byte("")), helm.InstallWait(true))
	if err != nil {
		t.Fatalf("could not install values configmap %v", err)
	}
	err = r.WaitForStatus(testConfigMapRelease, "DEPLOYED")
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deployed", cr))

	// Install Chartconfig
	l.Log("level", "debug", "message", fmt.Sprintf("creating %s", cr))
	chartValues, err := chartconfig.ExecuteChartValuesTemplate(chartConfigValues)
	if err != nil {
		t.Fatalf("could not template chart values %q %v", chartValues, err)
	}
	err = r.InstallResource(cr, chartValues, "stable")
	if err != nil {
		t.Fatalf("could not install %q %v", cr, err)
	}

	err = r.WaitForStatus(cr, "DEPLOYED")
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deployed", cr))

	err = r.WaitForStatus(testChartRelease, "DEPLOYED")
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deployed", testChartRelease))

	// Check if values are applied
	rc, err := helmClient.GetReleaseContent(testChartRelease)
	if err != nil {
		t.Fatalf("could not get release content of %s %v", testChartRelease, err)
	}
	l.Log("level", "debug", "message", fmt.Sprintf("chart %s has values %#v", testChartRelease, rc))
	if rc.Values["config"] == "config" {
		t.Fatalf("expected %#v got %#v", "config", rc.Values["config"])
	}
}
