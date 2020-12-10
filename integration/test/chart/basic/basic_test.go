// +build k8srequired

package basic

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/helmclient/v3/pkg/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/chart-operator/v2/integration/key"
)

const (
	configMapName = "test-app-configmap"
	secretName    = "test-app-configmap"

	configmapValue = `
v1:
  username: admin
  host:
    url: quay.io
  memory_in_gb: 0
  threshold: 2.17
replicas: 3
`
	secretValue = `
v1:
  username: admin
  host: 
    secret: 
      authToken: xer32wnq
  githubToken: nnbhwk1dk
  memory_in_gb: 3.14
`

	mergedValue = `
v1:
  username: admin
  host:
    url: quay.io
    secret: 
      authToken: xer32wnq
  githubToken: nnbhwk1dk
  memory_in_gb: 3.14
  threshold: 2.17
replicas: 3
`
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

	// creating dependant configmap & secret
	{
		config.Logger.Debugf(ctx, "creating configmap %#q", configMapName)

		cr := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      configMapName,
				Namespace: key.Namespace(),
			},
			Data: map[string]string{
				"values": configmapValue,
			},
		}

		_, err := config.K8sClients.K8sClient().CoreV1().ConfigMaps(key.Namespace()).Create(ctx, cr, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "creating configmap %#q", configMapName)

		config.Logger.Debugf(ctx, "creating secret %#q", secretName)

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: key.Namespace(),
			},
			StringData: map[string]string{
				"values": secretValue,
			},
		}

		_, err = config.K8sClients.K8sClient().CoreV1().Secrets(key.Namespace()).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.Debugf(ctx, "creating secret %#q", secretName)
	}

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
				Name:      key.TestAppReleaseName(),
				Namespace: key.Namespace(),
				Config: v1alpha1.ChartSpecConfig{
					ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
						Name:      configMapName,
						Namespace: key.Namespace(),
					},
					Secret: v1alpha1.ChartSpecConfigSecret{
						Name:      secretName,
						Namespace: key.Namespace(),
					},
				},
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

	// Check all values are merged correctly
	{
		config.Logger.Debugf(ctx, "comparing helm values %#q", key.TestAppReleaseName())

		content, err := config.HelmClient.GetReleaseContent(ctx, key.Namespace(), key.TestAppReleaseName())
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		var mergedMap map[string]interface{}
		{
			err = yaml.Unmarshal([]byte(mergedValue), &mergedMap)
			if err != nil {
				t.Fatalf("expected %#v got %#v", nil, err)
			}
		}

		if !reflect.DeepEqual(content.Values, mergedMap) {
			t.Fatalf("expected same got %s", cmp.Diff(content.Values, mergedMap))
		}

		config.Logger.Debugf(ctx, "compared helm values %#q", key.TestAppReleaseName())
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
