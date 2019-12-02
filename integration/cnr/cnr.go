// +build k8srequired

package cnr

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/k8sportforward"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Chart struct {
	Channel string
	Release string
	Tarball string
	Name    string
}

func Push(ctx context.Context, k8sClients *k8sclient.Clients, charts []Chart) error {
	var err error

	var forwarder *k8sportforward.Forwarder
	{
		c := k8sportforward.ForwarderConfig{
			RestConfig: k8sClients.RESTConfig(),
		}

		forwarder, err = k8sportforward.NewForwarder(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	podName, err := waitForPod(k8sClients, "giantswarm", "app=cnr-server")
	if err != nil {
		return microerror.Mask(err)
	}

	tunnel, err := forwarder.ForwardPort("giantswarm", podName, 5000)
	if err != nil {
		return microerror.Mask(err)
	}

	err = waitForServer("http://" + tunnel.LocalAddress() + "/cnr/api/v1/packages")
	if err != nil {
		return microerror.Mask(err)
	}

	l, err := micrologger.New(micrologger.Config{})
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
		err = a.PushChartTarball(ctx, chart.Name, chart.Release, chart.Tarball)
		if err != nil {
			return microerror.Mask(err)
		}

		err = a.PromoteChart(ctx, chart.Name, chart.Release, chart.Channel)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func waitForServer(url string) error {
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

func waitForPod(k8sClients *k8sclient.Clients, namespace, selector string) (string, error) {
	var err error
	var podName string

	operation := func() error {
		o := metav1.ListOptions{
			LabelSelector: selector,
		}
		pods, err := k8sClients.K8sClient().CoreV1().Pods(namespace).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(pods.Items) > 1 {
			return microerror.Maskf(executionFailedError, "expected 1 pod, found %d", len(pods.Items))
		}
		if len(pods.Items) == 0 {
			return microerror.Maskf(executionFailedError, "expected 1 pod, found 0")
		}
		podName = pods.Items[0].Name

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
