// +build k8srequired

package chart

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/k8sportforward"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
)

type Chart struct {
	Channel string
	Release string
	Tarball string
	Name    string
}

func Push(f *framework.Host, charts []Chart) error {
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
	for _, chart := range charts {
		err = a.PushChartTarball(chart.Name, chart.Release, chart.Tarball)
		if err != nil {
			return microerror.Mask(err)
		}

		err = a.PromoteChart(chart.Name, chart.Release, chart.Channel)
		if err != nil {
			return microerror.Mask(err)
		}
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
