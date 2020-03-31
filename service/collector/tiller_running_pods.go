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
	tillerRunningPodsDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "tiller_running_pods"),
		"Tiller running pods.",
		[]string{
			labelNamespace,
		},
		nil,
	)
)

// TillerRunningPodsConfig is this collector's configuration struct.
type TillerRunningPodsConfig struct {
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	Logger     micrologger.Logger

	TillerNamespace string
}

// TillerRunningPods is the main struct for this collector.
type TillerRunningPods struct {
	g8sClient  versioned.Interface
	helmClient helmclient.Interface
	logger     micrologger.Logger

	tillerNamespace string
}

// NewTillerRunningPods creates a new TillerRunningPods metrics collector.
func NewTillerRunningPods(config TillerRunningPodsConfig) (*TillerRunningPods, error) {
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

	t := &TillerRunningPods{
		g8sClient:  config.G8sClient,
		helmClient: config.HelmClient,
		logger:     config.Logger,

		tillerNamespace: config.TillerNamespace,
	}

	return t, nil
}

func (t *TillerRunningPods) Collect(ch chan<- prometheus.Metric) error {
	var value float64

	ctx := context.Background()

	charts, err := t.g8sClient.ApplicationV1alpha1().Charts("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	if len(charts.Items) == 0 {
		// Skip pinging Tiller when there are no chart CRs.
		// As Tiller is only installed when there is at least one CR to reconcile.
		t.logger.Log("level", "debug", "message", "did not collect Tiller running pods")
		t.logger.Log("level", "debug", "message", "no Chart or ChartConfig CRs in the cluster")

		value = 1
	} else {
		err := t.helmClient.PingTiller(ctx)
		if err != nil {
			t.logger.Log("level", "error", "message", "failed to collect Tiller running pods", "stack", fmt.Sprintf("%#v", err))

			value = 0
		} else {
			value = 1
		}
	}

	ch <- prometheus.MustNewConstMetric(
		tillerRunningPodsDesc,
		prometheus.GaugeValue,
		value,
		t.tillerNamespace,
	)

	return nil
}

// Describe emits the description for the metrics collected here.
func (t *TillerRunningPods) Describe(ch chan<- *prometheus.Desc) error {
	ch <- tillerRunningPodsDesc
	return nil
}
