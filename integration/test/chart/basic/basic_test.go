// +build k8srequired

package basic

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/chartvalues"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/integration/key"
)

// TestChartLifecycle tests a Helm release can be created, updated and deleted
// uaing a chart CR processed by chart-operator.
//
// - Create chart CR using apiextensions-chart-e2e-chart.
// - Ensure test app specfied in the chart CR is deployed.
//
// - Update chart CR using apiextensions-chart-e2e-chart.
// - Ensure test app is redeployed using updated chart tarball.
//
// - Delete apiextensions-chart-e2e-chart.
// - Ensure test app is deleted.
//
func TestChartLifecycle(t *testing.T) {
	ctx := context.Background()

	// Test creation.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating chart %#q", key.CustomResourceReleaseName()))

		c := chartvalues.APIExtensionsChartE2EConfig{
			Chart: chartvalues.APIExtensionsChartE2EConfigChart{
				Name:       key.TestAppReleaseName(),
				Namespace:  "giantswarm",
				TarballURL: "https://giantswarm.github.com/sample-catalog/kubernetes-test-app-chart-0.5.3.tgz",
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

		chartInfo := release.NewStableChartInfo(key.CustomResourceReleaseName())
		err = config.Release.Install(ctx, key.CustomResourceReleaseName(), chartInfo, chartValues)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = config.Release.WaitForStatus(ctx, fmt.Sprintf("%s-%s", "giantswarm", key.CustomResourceReleaseName()), "DEPLOYED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created chart %#q", key.CustomResourceReleaseName()))

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking release %#q is deployed", key.TestAppReleaseName()))

		err = config.Release.WaitForStatus(ctx, key.TestAppReleaseName(), "DEPLOYED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q is deployed", key.TestAppReleaseName()))
	}

	// Check chart CR status.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking status for chart CR %#q", key.TestAppReleaseName()))

		cr, err := config.K8sClients.G8sClient().ApplicationV1alpha1().Charts("giantswarm").Get(key.TestAppReleaseName(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		if cr.Status.Release.Status != "DEPLOYED" {
			t.Fatalf("expected CR release status %#q got %#q", "DEPLOYED", cr.Status.Release.Status)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked status for chart CR %#q", key.TestAppReleaseName()))
	}

	// Test update.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating chart %#q", key.CustomResourceReleaseName()))

		c := chartvalues.APIExtensionsChartE2EConfig{
			Chart: chartvalues.APIExtensionsChartE2EConfigChart{
				Name:      key.TestAppReleaseName(),
				Namespace: "giantswarm",
				// Newer version of the tarball is deployed.
				TarballURL: "https://giantswarm.github.com/sample-catalog/kubernetes-test-app-chart-0.6.8.tgz",
			},
			ChartOperator: chartvalues.APIExtensionsChartE2EConfigChartOperator{
				Version: "1.0.0",
			},
			Namespace: "giantswarm",
		}

		updatedValues, err := chartvalues.NewAPIExtensionsChartE2E(c)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		chartInfo := release.NewStableChartInfo(key.CustomResourceReleaseName())
		err = config.Release.Update(ctx, key.CustomResourceReleaseName(), chartInfo, updatedValues)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = config.Release.WaitForStatus(ctx, fmt.Sprintf("%s-%s", "giantswarm", key.CustomResourceReleaseName()), "DEPLOYED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated chart %#q", key.CustomResourceReleaseName()))

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking release %#q is updated", key.TestAppReleaseName()))

		err = config.Release.WaitForChartInfo(ctx, key.TestAppReleaseName(), "0.6.8")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q is updated", key.TestAppReleaseName()))
	}

	// Test deletion.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting chart %#q", key.CustomResourceReleaseName()))

		err := config.Release.Delete(ctx, key.CustomResourceReleaseName())
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = config.Release.WaitForStatus(ctx, fmt.Sprintf("%s-%s", "giantswarm", key.CustomResourceReleaseName()), "DELETED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted chart %#q", key.CustomResourceReleaseName()))

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking release %#q is deleted", key.TestAppReleaseName()))

		err = config.Release.WaitForStatus(ctx, key.TestAppReleaseName(), "DELETED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q is deleted", key.TestAppReleaseName()))
	}
}
