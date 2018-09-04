// +build k8srequired

package chartvalues

import (
	"fmt"
	"testing"

	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"

	"github.com/giantswarm/chart-operator/integration/chart"
	"github.com/giantswarm/chart-operator/integration/chartconfig"
	"github.com/giantswarm/chart-operator/integration/env"
)

func TestChartValues(t *testing.T) {
	const cr = "apiextensions-chart-config-e2e"

	charts := []chart.Chart{
		{
			Channel: "1-0-beta",
			Release: "1.0.0",
			Tarball: "/e2e/fixtures/tb-chart-1.0.0.tgz",
			Name:    "tb-chart",
		},
		{
			Channel: "1-0-beta",
			Release: "1.0.0",
			Tarball: "/e2e/fixtures/tb-configmap-1.0.0.tgz",
			Name:    "tb-configmap",
		},
	}

	chartConfigValues := e2etemplates.ApiextensionsChartConfigValues{
		Channel: "1-0-beta",
		ConfigMap: e2etemplates.ApiextensionsChartConfigConfigMap{
			Name:            "tb-configmap",
			Namespace:       "giantswarm",
			ResourceVersion: "1",
		},
		Name:                 "tb-chart",
		Namespace:            "giantswarm",
		Release:              "tb-release",
		VersionBundleVersion: env.VersionBundleVersion(),
	}
	err := chart.Push(h, charts)
	if err != nil {
		t.Fatalf("could not push inital charts to cnr %v", err)
	}

	// Test Creation

	// Install ValuesConfigMaps
	err = r.InstallResource("tb-configmap", "", "1-0-beta")
	if err != nil {
		t.Fatalf("could not install %q %v", cr, err)
	}

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
}
