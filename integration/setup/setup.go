// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/e2e-harness/pkg/release"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/integration/env"
	"github.com/giantswarm/chart-operator/integration/key"
	"github.com/giantswarm/chart-operator/integration/templates"
)

func Setup(m *testing.M, config Config) {
	ctx := context.Background()

	var v int
	var err error

	err = installResources(ctx, config)
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "error", "message", "failed to install resources", "stack", fmt.Sprintf("%#v", err))
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if env.KeepResources() != "true" {
		// only do full teardown when not on CI
		if env.CircleCI() != "true" {
			err := teardown(ctx, config)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}
		}
	}

	os.Exit(v)
}

func installResources(ctx context.Context, config Config) error {
	var err error

	{
		err = config.CPK8sSetup.EnsureNamespaceCreated(ctx, "giantswarm")
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = config.HelmClient.EnsureTillerInstalled(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = config.Release.InstallOperator(ctx, key.ChartOperatorReleaseName(), release.NewVersion(env.CircleSHA()), templates.ChartOperatorValues, v1alpha1.NewChartConfigCRD())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
