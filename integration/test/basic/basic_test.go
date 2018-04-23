// +build k8srequired

package basic

import (
	"log"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"k8s.io/helm/pkg/helm"
)

func TestChartInstalled(t *testing.T) {
	err := installChartOperatorResource(f, helmClient)
	if err != nil {
		t.Fatalf("could not install chart-operator-resource-chart %v", err)
	}

	var rc *helmclient.ReleaseContent
	operation := func() error {
		rc, err = helmClient.GetReleaseContent("tb-release")
		if err != nil {
			return microerror.Maskf(err, "could not retrieve release content")
		}
		if rc.Status == "PENDING_INSTALL" {
			return microerror.Newf("release still not installed")
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("waiting for release %s: %v", t, err)
	}

	err = backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
	if err != nil {
		t.Fatal("expected nil found", err)
	}

	expectedStatus := "DEPLOYED"
	if rc.Status != expectedStatus {
		t.Fatalf("unexpected chart status, want %q, got %q", expectedStatus, rc.Status)
	}
}

func installChartOperatorResource(f *framework.Host, helmClient *helmclient.Client) error {
	const chartOperatorResourceValues = `chart:
  name: "tb-chart"
  channel: "5-5-beta"
  namespace: "default"
  release: "tb-release"
`
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

	tarballPath, err := a.PullChartTarball("chart-operator-resource-chart", "stable")
	if err != nil {
		return microerror.Mask(err)
	}

	helmClient.InstallFromTarball(tarballPath, "kube-system",
		helm.ReleaseName("chart-operator-resource"),
		helm.ValueOverrides([]byte(chartOperatorResourceValues)),
		helm.InstallWait(true))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
