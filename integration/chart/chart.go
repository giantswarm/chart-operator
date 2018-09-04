// +build k8srequired

package chart

import (
	"fmt"
	"log"
	"net/http"
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

func Push(l micrologger.Logger, h *framework.Host, charts []Chart) error {
	var err error

	var forwarder *k8sportforward.Forwarder
	{
		c := k8sportforward.ForwarderConfig{
			RestConfig: h.RestConfig(),
		}

		forwarder, err = k8sportforward.NewForwarder(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	podName, err := waitForPod(h, "giantswarm", "app=cnr-server")
	if err != nil {
		return microerror.Mask(err)
	}

	tunnel, err := forwarder.ForwardPort("giantswarm", podName, 5000)
	if err != nil {
		return microerror.Mask(err)
	}

	err = waitForServer(h, "http://"+tunnel.LocalAddress()+"/cnr/api/v1/packages")
	if err != nil {
		return microerror.Mask(err)
	}

	c := apprclient.Config{
		Fs:     afero.NewOsFs(),
		Logger: l,

		Address:      "http://" + tunnel.LocalAddress(),
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

		l.Log("level", "debug", "message", fmt.Sprintf("pushed chart %s to channel %s", chart.Name, chart.Channel))
	}

	return nil
}

func waitForServer(h *framework.Host, url string) error {
	var err error

	operation := func() error {
		_, err := http.Get(url)
		if err != nil {
			return microerror.Mask(err)
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

func waitForPod(h *framework.Host, ns, selector string) (string, error) {
	var err error
	var podName string
	operation := func() error {
		podName, err = h.GetPodName(ns, selector)
		if err != nil {
			return microerror.Mask(err)
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
