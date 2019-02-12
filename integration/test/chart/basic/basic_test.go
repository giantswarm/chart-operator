// +build k8srequired

package basic

import (
	"context"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"

	"github.com/giantswarm/chart-operator/integration/chartconfig"
	"github.com/giantswarm/chart-operator/integration/key"
)

const (
	cr = "apiextensions-chart-e2e"
)

func TestChartLifecycle(t *testing.T) {
	ctx := context.Background()

	// Setup.
	err := chartconfig.InstallResources(ctx, config)
	if err != nil {
		t.Fatalf("could not install resources %v", err)
	}

	// Install chart CR.
	c := chartvalues.APIExtensionsChartE2EConfig{
		Chart: chartvalues.APIExtensionsChartE2EConfigChart{
			Name:       "kubernetes-test-app",
			Namespace:  "giantswarm",
			TarballURL: "https://giantswarm.github.com/sample-catalog/kubernetes-test-app-chart-0.6.8.tgz",
		},
		ChartOperator: chartvalues.APIExtensionsChartE2EConfigChartOperator{
			Version: "1.0.0",
		},
		Namespace: "giantswarm",
	}

	values, err := chartvalues.NewAPIExtensionsChartE2E(c)
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}

	chartInfo := release.NewStableChartInfo(key.ChartCustomResource())
	err = config.Release.EnsureInstalled(ctx, key.ChartCustomResource(), chartInfo, values)
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}
}
