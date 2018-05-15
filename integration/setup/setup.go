// +build k8srequired

package setup

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/k8sportforward"
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

	err = installInitialCharts(f)
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

func installInitialCharts(f *framework.Host) error {
	fwc := k8sportforward.Config{
		RestConfig: f.RestConfig(),
	}

	fw, err := k8sportforward.New(fwc)
	if err != nil {
		return microerror.Mask(err)
	}

	podName, err := waitForPod(f, "giantswarm", "app=cnr-server")
	if err != nil {
		return microerror.Mask(err)
	}
	tc := k8sportforward.TunnelConfig{
		Remote:    5000,
		Namespace: "giantswarm",
		PodName:   podName,
	}
	tunnel, err := fw.ForwardPort(tc)
	if err != nil {
		return microerror.Mask(err)
	}

	serverAddress := "http://localhost:" + strconv.Itoa(tunnel.Local)
	err = waitForServer(f, serverAddress+"/cnr/api/v1/packages")
	if err != nil {
		return microerror.Mask(fmt.Errorf("server didn't come up on time"))
	}

	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		return microerror.Mask(err)
	}

	c := apprclient.Config{
		Fs:     afero.NewOsFs(),
		Logger: l,

		Address:      serverAddress,
		Organization: "giantswarm",
	}

	a, err := apprclient.New(c)
	if err != nil {
		return microerror.Mask(err)
	}

	err = a.PushChartTarball("tb-chart", "5.5.5", "/e2e/fixtures/tb-chart-5.5.5.tgz")
	if err != nil {
		return microerror.Mask(err)
	}

	err = a.PromoteChart("tb-chart", "5.5.5", "5-5-beta")
	if err != nil {
		return microerror.Mask(err)
	}

	err = a.PushChartTarball("tb-chart", "5.6.0", "/e2e/fixtures/tb-chart-5.6.0.tgz")
	if err != nil {
		return microerror.Mask(err)
	}

	err = a.PromoteChart("tb-chart", "5.6.0", "5-6-beta")
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func waitForServer(f *framework.Host, url string) error {
	var err error

	operation := func() error {
		_, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("could not retrieve %s: %v", url, err)
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("waiting for server at %s: %v", t, err)
	}

	err = backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func waitForPod(f *framework.Host, ns, selector string) (string, error) {
	var err error
	var podName string
	operation := func() error {
		podName, err = f.GetPodName(ns, selector)
		if err != nil {
			return fmt.Errorf("could not retrieve pod %q on %q: %v", selector, ns, err)
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		log.Printf("waiting for pod at %s: %v", t, err)
	}

	err = backoff.RetryNotify(operation, backoff.NewExponentialBackOff(), notify)
	if err != nil {
		return "", microerror.Mask(err)
	}
	return podName, nil
}
