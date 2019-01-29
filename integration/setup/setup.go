// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/integration/env"
	"github.com/giantswarm/chart-operator/integration/teardown"
	"github.com/giantswarm/chart-operator/integration/templates"
)

func WrapTestMain(ctx context.Context, h *framework.Host, helmClient *helmclient.Client, l micrologger.Logger, m *testing.M) {
	var v int
	var err error

	err = h.CreateNamespace("giantswarm")
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = helmClient.EnsureTillerInstalled(ctx)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	if env.TestedCustomResource() == env.ChartConfigCustomResource {
		err = chartConfigResources(ctx, h, helmClient, l)
		if err != nil {
			log.Printf("%#v\n", err)
			v = 1
		}
	}

	if v == 0 {
		v = m.Run()
	}

	if env.KeepResources() != "true" {
		// only do full teardown when not on CI
		if env.CircleCI() != "true" {
			err := teardown.Teardown(h, helmClient)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			h.Teardown()
		}
	}

	os.Exit(v)
}

func chartConfigResources(ctx context.Context, h *framework.Host, helmClient *helmclient.Client, l micrologger.Logger) error {
	err := initializeCNR(ctx, h, helmClient, l)
	if err != nil {
		return microerror.Mask(err)
	}

	version := fmt.Sprintf(":%s", env.CircleSHA())
	err = h.InstallOperator("chart-operator", "chartconfig", templates.ChartOperatorValues, version)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func initializeCNR(ctx context.Context, h *framework.Host, helmClient *helmclient.Client, l micrologger.Logger) error {
	err := installCNR(ctx, h, helmClient, l)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installCNR(ctx context.Context, h *framework.Host, helmClient *helmclient.Client, l micrologger.Logger) error {
	c := apprclient.Config{
		Fs:     afero.NewOsFs(),
		Logger: l,

		Address:      "https://quay.io",
		Organization: "giantswarm",
	}

	a, err := apprclient.New(c)
	if err != nil {
		return microerror.Mask(err)
	}

	tarball, err := a.PullChartTarball(ctx, "cnr-server-chart", "stable")
	if err != nil {
		return microerror.Mask(err)
	}

	err = helmClient.InstallReleaseFromTarball(context.TODO(), tarball, "giantswarm", helm.ReleaseName("cnr-server"), helm.ValueOverrides([]byte("{}")), helm.InstallWait(true))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
