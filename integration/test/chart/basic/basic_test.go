// +build k8srequired

package basic

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/apiextensions/v2/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v2/pkg/label"
	"github.com/giantswarm/helmclient/v2/pkg/helmclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/v2/integration/key"
)

// TestChartLifecycle tests a Helm release can be created, updated and deleted
// uaing a chart CR processed by chart-operator.
//
// - Create chart CR.
// - Ensure test app specified in the chart CR is deployed.
//
// - Update chart CR.
// - Ensure test app is redeployed using updated chart tarball.
//
// - Delete chart CR.
// - Ensure test app is deleted.
//
func TestChartLifecycle(t *testing.T) {
	ctx := context.Background()

	// Test creation.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating chart %#q", key.TestAppReleaseName()))

		cr := &v1alpha1.Chart{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.TestAppReleaseName(),
				Namespace: key.Namespace(),
				Labels: map[string]string{
					label.ChartOperatorVersion: "1.0.0",
				},
			},
			Spec: v1alpha1.ChartSpec{
				Name:       key.TestAppReleaseName(),
				Namespace:  key.Namespace(),
				TarballURL: "https://giantswarm.github.io/default-catalog/test-app-0.1.0.tgz",
				Version:    "0.1.0",
			},
		}
		_, err := config.K8sClients.G8sClient().ApplicationV1alpha1().Charts(key.Namespace()).Create(ctx, cr, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created chart %#q", key.TestAppReleaseName()))

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking release %#q is deployed", key.TestAppReleaseName()))

		err = config.Release.WaitForStatus(ctx, key.Namespace(), key.TestAppReleaseName(), helmclient.StatusDeployed)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q is deployed", key.TestAppReleaseName()))
	}

	// Check chart CR status.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking status for chart CR %#q", key.TestAppReleaseName()))

		cr, err := config.K8sClients.G8sClient().ApplicationV1alpha1().Charts(key.Namespace()).Get(ctx, key.TestAppReleaseName(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		if cr.Status.Release.Status != helmclient.StatusDeployed {
			t.Fatalf("expected CR release status %#q got %#q", helmclient.StatusDeployed, cr.Status.Release.Status)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checked status for chart CR %#q", key.TestAppReleaseName()))
	}

	// Test update.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating chart %#q", key.TestAppReleaseName()))

		cr, err := config.K8sClients.G8sClient().ApplicationV1alpha1().Charts(key.Namespace()).Get(ctx, key.TestAppReleaseName(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		cr.Spec.TarballURL = "https://giantswarm.github.io/default-catalog/test-app-0.1.1.tgz"
		cr.Spec.Version = "0.1.1"

		_, err = config.K8sClients.G8sClient().ApplicationV1alpha1().Charts(key.Namespace()).Update(ctx, cr, metav1.UpdateOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated chart %#q", key.TestAppReleaseName()))

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking release %#q is updated", key.TestAppReleaseName()))

		err = config.Release.WaitForChartVersion(ctx, key.Namespace(), key.TestAppReleaseName(), "0.1.1")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q is updated", key.TestAppReleaseName()))
	}

	// Test deletion.
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting chart %#q", key.TestAppReleaseName()))

		err := config.K8sClients.G8sClient().ApplicationV1alpha1().Charts(key.Namespace()).Delete(ctx, key.TestAppReleaseName(), metav1.DeleteOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted chart %#q", key.TestAppReleaseName()))

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("checking release %#q is deleted", key.TestAppReleaseName()))

		err = config.Release.WaitForStatus(ctx, key.Namespace(), key.TestAppReleaseName(), helmclient.StatusUninstalled)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q is deleted", key.TestAppReleaseName()))
	}
}
