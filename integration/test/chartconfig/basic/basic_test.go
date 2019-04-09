// +build k8srequired

package basic

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/e2etemplates/pkg/e2etemplates"
	"github.com/giantswarm/microerror"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/chart-operator/integration/chartconfig"
	"github.com/giantswarm/chart-operator/integration/cnr"
	"github.com/giantswarm/chart-operator/integration/env"
	"github.com/giantswarm/chart-operator/integration/setup"
)

const (
	crName      = "tb-chart"
	crNamespace = "default"
	releaseName = "tb-release"
)

func TestChartLifecycle(t *testing.T) {
	ctx := context.Background()

	// Setup
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

		err := cnr.Push(ctx, config.Host, charts)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	versionBundleVersion, err := chartconfig.VersionBundleVersion(env.GithubToken(), env.TestedVersion())
	if err != nil {
		t.Fatalf("could not get version bundle version %v", err)
	}

	// Test Creation

	chartConfigValuesParams := e2etemplates.ApiextensionsChartConfigValues{
		Channel:              "5-5-beta",
		Name:                 crName,
		Namespace:            crNamespace,
		Release:              "tb-release",
		VersionBundleVersion: versionBundleVersion,
	}

	{
		values, err := chartconfig.ExecuteValuesTemplate(chartConfigValuesParams)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating ChartConfig %#q in namespace %#q", crName, crNamespace))

		chartValues, err := chartconfig.ExecuteValuesTemplate(chartConfigValuesParams)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = config.Release.EnsureInstalled(ctx, releaseName, release.NewStableChartInfo("apiextensions-chart-config-e2e-chart"), values, crExistsCondition(ctx, config, corev1alpha1.NewChartConfigCRD(), crNamespace, crName))
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)

		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created ChartConfig %#q in namespace %#q", crName, crNamespace))
	}

	// Test Update
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating ChartConfig %#q in namespace %#q", crName, crNamespace))

		chartConfigValuesParams.Channel = "5-6-beta"
		values, err := chartconfig.ExecuteValuesTemplate(chartConfigValuesParams)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		err = config.Release.Update(ctx, releaseName, release.NewStableChartInfo("apiextensions-chart-config-e2e-chart"), values, crExistsCondition(ctx, config, corev1alpha1.NewChartConfigCRD(), crNamespace, crName))
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		// TODO port WaitForVersion to e2e-harness/pkg/release.Release
		err = config.Resource.WaitForVersion(releaseName, "5.6.0")
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated ChartConfig %#q in namespace %#q", crName, crNamespace))
	}

	// Test Deletion
	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting ChartConfig %#q in namespace %#q", crName, crNamespace))

		err = config.Release.EnsureDeleted(ctx, releaseName, crNotFoundCondition(ctx, config, corev1alpha1.NewChartConfigCRD(), crNamespace, crName))
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted ChartConfig %#q in namespace %#q", crName, crNamespace))
	}
}

func crExistsCondition(ctx context.Context, config setup.Config, crd *apiextensionsv1beta1.CustomResourceDefinition, crNamespace, crName string) release.ConditionFunc {
	return func() error {
		gvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  crd.Spec.Version,
			Resource: crd.Spec.Names.Plural,
		}

		var dynamicClient dynamic.Interface
		{
			var err error

			dynamicClient, err = dynamic.NewForConfig(rest.CopyConfig(config.Host.RestConfig()))
			if err != nil {
				return microerror.Mask(err)
			}
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for creation of CR %#q in namespace %#q", crName, crNamespace))

		o := func() error {
			_, err := dynamicClient.Resource(gvr).Namespace(crNamespace).Get(crName, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return microerror.Maskf(notFoundError, "CR %#q in namespace %#q", crName, crNamespace)
			} else if err != nil {
				return backoff.Permanent(microerror.Mask(err))
			}

			return nil
		}
		b := backoff.NewExponential(5*time.Minute, 1*time.Minute)
		n := backoff.NewNotifier(config.Logger, ctx)

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for creation of CR %#q in namespace %#q", crName, crNamespace))
		return nil
	}
}

func crNotFoundCondition(ctx context.Context, config setup.Config, crd *apiextensionsv1beta1.CustomResourceDefinition, crNamespace, crName string) release.ConditionFunc {
	return func() error {
		gvr := schema.GroupVersionResource{
			Group:    crd.Spec.Group,
			Version:  crd.Spec.Version,
			Resource: crd.Spec.Names.Plural,
		}

		var dynamicClient dynamic.Interface
		{
			var err error

			dynamicClient, err = dynamic.NewForConfig(rest.CopyConfig(config.Host.RestConfig()))
			if err != nil {
				return microerror.Mask(err)
			}
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waiting for deletion of CR %#q in namespace %#q", crName, crNamespace))

		o := func() error {
			_, err := dynamicClient.Resource(gvr).Namespace(crNamespace).Get(crName, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return nil
			} else if err != nil {
				return backoff.Permanent(microerror.Mask(err))
			}

			return microerror.Maskf(stillExistsError, "CR %#q in namespace %#q", crName, crNamespace)
		}
		b := backoff.NewExponential(60*time.Minute, 5*time.Minute)
		n := backoff.NewNotifier(config.Logger, ctx)

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("waited for deletion of CR %#q in namespace %#q", crName, crNamespace))
		return nil
	}
}
