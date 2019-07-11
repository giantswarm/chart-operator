package collector

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

var (
	chartCordonExpireTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "chart", "cordon_expire_time_seconds"),
		"A metric of the expire time of cordoned charts as unix seconds.",
		[]string{
			labelRelease,
		},
		nil,
	)
)

// ChartResourceConfig is this collector's configuration struct.
type ChartResourceConfig struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

// ChartResource is the main struct for this collector.
type ChartResource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

// NewChartResource creates a new ChartResource metrics collector.
func NewChartResource(config ChartResourceConfig) (*ChartResource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	a := &ChartResource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return a, nil
}

func (c *ChartResource) Collect(ch chan<- prometheus.Metric) error {
	c.logger.Log("level", "debug", "message", "collecting metrics for Charts")

	charts, err := c.g8sClient.ApplicationV1alpha1().Charts("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, chart := range charts.Items {

		if !key.IsCordoned(chart) {
			continue
		}

		t, err := convertToTime(key.CordonUntil(chart))
		if err != nil {
			c.logger.Log("level", "warning", "message", "could not convert cordon-until", "stack", fmt.Sprintf("%#v", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			chartCordonExpireTimeDesc,
			prometheus.GaugeValue,
			float64(t.Unix()),
			key.ReleaseName(chart),
		)
	}

	c.logger.Log("level", "debug", "message", "finished collecting metrics for Charts")

	return nil
}

// Describe emits the description for the metrics collected here.
func (a *ChartResource) Describe(ch chan<- *prometheus.Desc) error {
	ch <- chartCordonExpireTimeDesc
	return nil
}
