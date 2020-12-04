// +build k8srequired

package basic

import (
	"context"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/helmclient/v3/pkg/helmclient"
	"github.com/giantswarm/microerror"
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
		config.Logger.Debugf(ctx, "creating chart %#q", key.TestAppReleaseName())

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

		config.Logger.Debugf(ctx, "created chart %#q", key.TestAppReleaseName())

		config.Logger.Debugf(ctx, "checking release %#q is deployed", key.TestAppReleaseName())

		err = config.Release.WaitForStatus(ctx, key.Namespace(), key.TestAppReleaseName(), helmclient.StatusDeployed)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "release %#q is deployed", key.TestAppReleaseName())
	}

	// Check chart CR status.
	{
		config.Logger.Debugf(ctx, "checking status for chart CR %#q", key.TestAppReleaseName())

		operation := func() error {
			cr, err := config.K8sClients.G8sClient().ApplicationV1alpha1().Charts(key.Namespace()).Get(ctx, key.TestAppReleaseName(), metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			if cr.Status.Release.Status != helmclient.StatusDeployed {
				return microerror.Mask(notDeployedError)
			}
			return nil
		}

		b := backoff.NewMaxRetries(10, 3*time.Second)
		err := backoff.Retry(operation, b)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "checked status for chart CR %#q", key.TestAppReleaseName())
	}

	// Test update.
	{
		config.Logger.Debugf(ctx, "updating chart %#q", key.TestAppReleaseName())

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

		config.Logger.Debugf(ctx, "updated chart %#q", key.TestAppReleaseName())

		config.Logger.Debugf(ctx, "checking release %#q is updated", key.TestAppReleaseName())

		err = config.Release.WaitForChartVersion(ctx, key.Namespace(), key.TestAppReleaseName(), "0.1.1")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "release %#q is updated", key.TestAppReleaseName())
	}

	// Test deletion.
	{
		config.Logger.Debugf(ctx, "deleting chart %#q", key.TestAppReleaseName())

		err := config.K8sClients.G8sClient().ApplicationV1alpha1().Charts(key.Namespace()).Delete(ctx, key.TestAppReleaseName(), metav1.DeleteOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "deleted chart %#q", key.TestAppReleaseName())

		config.Logger.Debugf(ctx, "checking release %#q is deleted", key.TestAppReleaseName())

		err = config.Release.WaitForStatus(ctx, key.Namespace(), key.TestAppReleaseName(), helmclient.StatusUninstalled)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "release %#q is deleted", key.TestAppReleaseName())
	}
}
