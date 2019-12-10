// +build k8srequired

package basic

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"

	"github.com/giantswarm/chart-operator/integration/chartconfig"
	"github.com/giantswarm/chart-operator/integration/cnr"
	"github.com/giantswarm/chart-operator/integration/env"
	"github.com/giantswarm/chart-operator/integration/key"
)

const (
	namespace   = "giantswarm"
	testRelease = "tb-release"
)

func TestChartLifecycle(t *testing.T) {
	ctx := context.Background()

	// Setup
	cr := key.ChartConfigReleaseName()
	chartInfo := release.NewStableChartInfo(cr)
	err := chartconfig.InstallResources(ctx, config)
	if err != nil {
		t.Fatalf("could not install resources %v", err)
	}

	{
		charts := []cnr.Chart{
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

		err := cnr.Push(ctx, config.K8sClients, charts)
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

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating %#q", cr))
		chartValues, err := chartconfig.ExecuteValuesTemplate(chartConfigValues)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = config.Release.Install(ctx, cr, chartInfo, chartValues)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = config.Release.WaitForStatus(ctx, fmt.Sprintf("%s-%s", namespace, cr), "DEPLOYED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%#q succesfully deployed", cr))

		err = config.Release.WaitForStatus(ctx, testRelease, "DEPLOYED")
		if err != nil {
			err = config.Release.WaitForStatus(ctx, testRelease, "DEPLOYED")
			if err != nil {
				t.Fatalf("expected %#v got %#v", nil, err)
			}
		}
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%#q succesfully deployed", testRelease))
	}

	// Test Update
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating %#q", cr))
		chartConfigValues.Channel = "5-6-beta"
		chartValues, err := chartconfig.ExecuteValuesTemplate(chartConfigValues)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = config.Release.Update(ctx, cr, chartInfo, chartValues)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = config.Release.WaitForChartInfo(ctx, testRelease, "5.6.0")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%#q successfully updated", testRelease))
	}

	// Test Deletion
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting %#q", cr))
		err := config.Release.Delete(ctx, cr)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = config.Release.WaitForStatus(ctx, testRelease, "DELETED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("%#q successfully deleted", testRelease))
	}
}
