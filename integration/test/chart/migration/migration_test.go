// +build k8srequired

package migration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/integration/key"
	"github.com/giantswarm/chart-operator/pkg/annotation"
)

// TestChartMigration tests chartconfig CR is deleted once it has been migrated
// to a chart CR. It simulates the migration steps performed by cluster-operator.
//
// Create chartconfig CRD.
// Create chartconfig CR.
// Create chart CR.
//
// Ensure test-app is deployed.
//
// Add annotation to chartconfig CR to mark that migration is complete.
// Delete chartconfig CR.
//
// Ensure that finalizer is removed and chartconfig CR is deleted.
// Ensure test-app is still deployed.
//
func TestChartMigration(t *testing.T) {
	ctx := context.Background()

	// Create legacy chartconfig CRD.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "creating chartconfig CRD")

		err := config.K8sClients.CRDClient().EnsureCreated(ctx, corev1alpha1.NewChartConfigCRD(), backoff.NewMaxRetries(7, 1*time.Second))
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "created chartconfig CRD")
	}

	// Create legacy chartconfig CR.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating chartconfig %#q", key.TestAppReleaseName()))

		chartConfig := &corev1alpha1.ChartConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.TestAppReleaseName(),
				Namespace: "giantswarm",
				Finalizers: []string{
					// Finalizer is created manually because there is no
					// chartconfig controller.
					"operatorkit.giantswarm.io/chart-operator-chartconfig",
				},
				Labels: map[string]string{
					"app": "test-app",
				},
			},
			Spec: corev1alpha1.ChartConfigSpec{
				Chart: corev1alpha1.ChartConfigSpecChart{
					Channel:   "0-7-beta",
					Name:      key.TestAppReleaseName(),
					Namespace: "giantswarm",
					Release:   key.TestAppReleaseName(),
				},
				VersionBundle: corev1alpha1.ChartConfigSpecVersionBundle{
					Version: "0.7.0",
				},
			},
		}
		_, err := config.K8sClients.G8sClient().CoreV1alpha1().ChartConfigs("giantswarm").Create(chartConfig)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created chartconfig %#q", key.TestAppReleaseName()))
	}

	// Create chart CR.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating chart %#q", key.TestAppReleaseName()))

		chart := &v1alpha1.Chart{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.TestAppReleaseName(),
				Namespace: "giantswarm",
				Labels: map[string]string{
					"app":                                  "test-app",
					"chart-operator.giantswarm.io/version": "1.0.0",
				},
			},
			Spec: v1alpha1.ChartSpec{
				Name:       key.TestAppReleaseName(),
				Namespace:  "giantswarm",
				TarballURL: "https://giantswarm.github.com/sample-catalog/kubernetes-test-app-chart-0.7.0.tgz",
			},
		}
		_, err := config.K8sClients.G8sClient().ApplicationV1alpha1().Charts("giantswarm").Create(chart)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created chart %#q", key.TestAppReleaseName()))
	}

	// Check test-app is deployed.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking release %#q is deployed", key.TestAppReleaseName()))

		err := config.Release.WaitForStatus(ctx, key.TestAppReleaseName(), "DEPLOYED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q is deployed", key.TestAppReleaseName()))
	}

	// Add annotation to chartconfig CR to mark that migration is complete.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("adding annotation to chartconfig %#q", key.TestAppReleaseName()))

		chartConfig, err := config.K8sClients.G8sClient().CoreV1alpha1().ChartConfigs("giantswarm").Get(key.TestAppReleaseName(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		annotations := map[string]string{}

		if chartConfig.Annotations != nil && len(chartConfig.Annotations) > 0 {
			annotations = chartConfig.Annotations
		}

		annotations[annotation.DeleteCustomResourceOnly] = "true"
		chartConfig.Annotations = annotations

		_, err = config.K8sClients.G8sClient().CoreV1alpha1().ChartConfigs("giantswarm").Update(chartConfig)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("added annotation to chartconfig %#q", key.TestAppReleaseName()))
	}

	// Delete chartconfig CR.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting chartconfig %#q", key.TestAppReleaseName()))

		err := config.K8sClients.G8sClient().CoreV1alpha1().ChartConfigs("giantswarm").Delete(key.TestAppReleaseName(), &metav1.DeleteOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted chartconfig %#q", key.TestAppReleaseName()))
	}

	// Check chartconfig CR is deleted.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking chartconfig %#q was deleted", key.TestAppReleaseName()))

		o := func() error {
			_, err := config.K8sClients.G8sClient().CoreV1alpha1().ChartConfigs("giantswarm").Get(key.TestAppReleaseName(), metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				// Error is expected because finalizer was removed.
				return nil
			} else if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}

		n := func(err error, t time.Duration) {
			config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("failed to get not found error: retrying in %s", t), "stack", fmt.Sprintf("%v", err))
		}

		b := backoff.NewExponential(backoff.MediumMaxWait, backoff.LongMaxInterval)
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked chartconfig %#q was deleted", key.TestAppReleaseName()))
	}

	// Check test-app is still deployed.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking release %#q is deployed", key.TestAppReleaseName()))

		err := config.Release.WaitForStatus(ctx, key.TestAppReleaseName(), "DEPLOYED")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q is deployed", key.TestAppReleaseName()))
	}
}
