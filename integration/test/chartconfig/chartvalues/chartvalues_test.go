// +build k8srequired

package chartvalues

import (
	"context"
	"fmt"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"testing"

	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"

	"github.com/giantswarm/chart-operator/integration/chartconfig"
	"github.com/giantswarm/chart-operator/integration/cnr"
	"github.com/giantswarm/chart-operator/integration/env"
)

func TestChartValues(t *testing.T) {
	const cr = "apiextensions-chart-config-e2e"

	ctx := context.Background()

	err := chartconfig.InstallResources(ctx, config)
	if err != nil {
		t.Fatalf("could not install resources %v", err)
	}

	charts := []cnr.Chart{
		{
			Channel: "1-0-beta",
			Release: "1.0.0",
			Tarball: "/e2e/fixtures/tb-chart-1.0.0.tgz",
			Name:    "tb-chart",
		},
	}

	chartConfigValues := e2etemplates.ApiextensionsChartConfigValues{
		Channel:              "1-0-beta",
		Name:                 "tb-chart",
		Namespace:            "giantswarm",
		Release:              "tb-release",
		VersionBundleVersion: env.VersionBundleVersion(),
	}
	err = cnr.Push(ctx, config.K8sClients, charts)
	if err != nil {
		t.Fatalf("could not push inital charts to cnr %v", err)
	}

	// Test Creation
	config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating %#q", cr))
	chartValues, err := chartconfig.ExecuteValuesTemplate(chartConfigValues)
	if err != nil {
		t.Fatalf("could not template chart values %q %v", chartValues, err)
	}

	chartInfo := release.NewStableChartInfo(cr)

	err = config.Release.Install(ctx, cr, chartInfo, chartValues)
	if err != nil {
		t.Fatalf("could not install %q %v", cr, err)
	}

	err = chartconfig.DeleteResources(ctx, config)
	if err != nil {
		t.Fatalf("could not delete resources %v", err)
	}
}
