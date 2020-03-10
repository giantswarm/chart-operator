package collector

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	tillerReachableDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "tiller_reachable"),
		"Tiller is reachable from chart-operator.",
		[]string{
			labelNamespace,
		},
		nil,
	)
)

// TillerReachableConfig is this collector's configuration struct.
type TillerReachableConfig struct {
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	Logger     micrologger.Logger

	TillerNamespace string
}

// TillerReachable is the main struct for this collector.
type TillerReachable struct {
	g8sClient  versioned.Interface
	helmClient helmclient.Interface
	logger     micrologger.Logger

	tillerNamespace string
}

// NewTillerReachable creates a new TillerReachable metrics collector.
func NewTillerReachable(config TillerReachableConfig) (*TillerReachable, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.TillerNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TillerNamespace must not be empty", config)
	}

	t := &TillerReachable{
		g8sClient:  config.G8sClient,
		helmClient: config.HelmClient,
		logger:     config.Logger,

		tillerNamespace: config.TillerNamespace,
	}

	return t, nil
}

func (t *TillerReachable) Collect(ch chan<- prometheus.Metric) error {
	var value float64

	ctx := context.Background()

	charts, err := t.g8sClient.ApplicationV1alpha1().Charts("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	if len(charts.Items) == 0 {
		// Skip pinging tiller when there are no chart CRs,
		// As Tiller is only installed when there is at least one CR to reconcile.
		t.logger.Log("level", "debug", "message", "did not collect Tiller reachability")
		t.logger.Log("level", "debug", "message", "no Chart or ChartConfig CRs in the cluster")

		value = 1
	} else {
		err := t.helmClient.PingTiller(ctx)
		if err != nil {
			t.logger.Log("level", "error", "message", "failed to collect Tiller reachability", "stack", fmt.Sprintf("%#v", err))

			value = 0
		} else {
			value = 1
		}
	}

	ch <- prometheus.MustNewConstMetric(
		tillerReachableDesc,
		prometheus.GaugeValue,
		value,
		t.tillerNamespace,
	)

	return nil
}

// Describe emits the description for the metrics collected here.
func (t *TillerReachable) Describe(ch chan<- *prometheus.Desc) error {
	ch <- tillerReachableDesc
	return nil
}
