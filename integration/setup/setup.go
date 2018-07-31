// +build k8srequired

package setup

import (
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

	"github.com/giantswarm/chart-operator/integration/teardown"
	"github.com/giantswarm/chart-operator/integration/templates"
)

func WrapTestMain(f *framework.Host, helmClient *helmclient.Client, m *testing.M) {
	var v int
	var err error

	err = f.CreateNamespace("giantswarm")
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = resources(f, helmClient)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
		// only do full teardown when not on CI
		if os.Getenv("CIRCLECI") != "true" {
			err := teardown.Teardown(f, helmClient)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			f.Teardown()
		}
	}

	os.Exit(v)
}

func resources(f *framework.Host, helmClient *helmclient.Client) error {
	err := initializeCNR(f, helmClient)
	if err != nil {
		return microerror.Mask(err)
	}

	err = f.InstallOperator("chart-operator", "chartconfig", templates.ChartOperatorValues, ":${CIRCLE_SHA1}")

	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func initializeCNR(f *framework.Host, helmClient *helmclient.Client) error {
	err := installCNR(f, helmClient)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installCNR(f *framework.Host, helmClient *helmclient.Client) error {
	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		return microerror.Mask(err)
	}

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

	tarball, err := a.PullChartTarball("cnr-server-chart", "stable")
	if err != nil {
		return microerror.Mask(err)
	}

	err = helmClient.InstallFromTarball(tarball, "giantswarm",
		helm.ReleaseName("cnr-server"),
		helm.ValueOverrides([]byte("{}")),
		helm.InstallWait(true))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
