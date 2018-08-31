// +build k8srequired

package basic

import (
	"fmt"
	"testing"

	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"

	"github.com/giantswarm/chart-operator/integration/chart"
	"github.com/giantswarm/chart-operator/integration/chartconfig"
	"github.com/giantswarm/chart-operator/integration/env"
)

const (
	testRelease = "tb-release"
	cr          = "apiextensions-chart-config-e2e"
)

func TestChartLifecycle(t *testing.T) {
	// Setup
	{
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

		err := chart.Push(h, charts)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	// Test Creation
	var chartConfigValues e2etemplates.ApiextensionsChartConfigValues
	{
		chartConfigValues = e2etemplates.ApiextensionsChartConfigValues{
			Channel:              "5-5-beta",
			Name:                 "tb-chart",
			Namespace:            "giantswarm",
			Release:              "tb-release",
			VersionBundleVersion: env.VersionBundleVersion(),
		}

		l.Log("level", "debug", "message", fmt.Sprintf("creating %s", cr))
		chartValues, err := chartconfig.ExecuteChartValuesTemplate(chartConfigValues)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = r.InstallResource(cr, chartValues, "stable")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = r.WaitForStatus(cr, "DEPLOYED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deployed", cr))

		err = r.WaitForStatus(testRelease, "DEPLOYED")
		if err != nil {
			err = r.WaitForStatus(testRelease, "DEPLOYED")
			if err != nil {
				t.Fatalf("expected %#v got %#v", nil, err)
			}
		}
		l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deployed", testRelease))
	}

	// Test Update
	{
		l.Log("level", "debug", "message", fmt.Sprintf("updating %s", cr))
		chartConfigValues.Channel = "5-6-beta"
		chartValues, err := chartconfig.ExecuteChartValuesTemplate(chartConfigValues)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		err = r.UpdateResource(cr, chartValues, "stable")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = r.WaitForVersion(testRelease, "5.6.0")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully updated", testRelease))
	}

	// Test Deletion
	{
		l.Log("level", "debug", "message", fmt.Sprintf("deleting %s", cr))
		err := helmClient.DeleteRelease(cr)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = r.WaitForStatus(testRelease, "DELETED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		l.Log("level", "debug", "message", fmt.Sprintf("%s succesfully deleted", testRelease))
	}
}
