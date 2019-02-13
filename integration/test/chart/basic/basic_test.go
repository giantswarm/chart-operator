// +build k8srequired

package basic

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"

	"github.com/giantswarm/chart-operator/integration/key"
)

func TestChartLifecycle(t *testing.T) {
	ctx := context.Background()

	// Test creation.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating chart %#q", key.ChartCustomResource()))

		c := chartvalues.APIExtensionsChartE2EConfig{
			Chart: chartvalues.APIExtensionsChartE2EConfigChart{
				Name:       key.TestApp(),
				Namespace:  "giantswarm",
				TarballURL: "https://giantswarm.github.com/sample-catalog/kubernetes-test-app-chart-0.6.8.tgz",
			},
			ChartOperator: chartvalues.APIExtensionsChartE2EConfigChartOperator{
				Version: "1.0.0",
			},
			Namespace: "giantswarm",
		}

		chartValues, err := chartvalues.NewAPIExtensionsChartE2E(c)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		chartInfo := release.NewStableChartInfo(key.ChartCustomResource())
		err = config.Release.Install(ctx, key.ChartCustomResource(), chartInfo, chartValues)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = config.Release.WaitForStatus(ctx, fmt.Sprintf("%s-%s", "giantswarm", key.ChartCustomResource()), "DEPLOYED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created chart %#q", key.ChartCustomResource()))

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking release %#q is deployed", key.TestApp()))

		err = config.Release.WaitForStatus(ctx, key.TestApp(), "DEPLOYED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q is deployed", key.TestApp()))
	}
}
