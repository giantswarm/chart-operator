package collector

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v7/key"
)

const (
	chartNameLabel     = "chart_name"
	channelNameLabel   = "channel_name"
	releaseNameLabel   = "release_name"
	releaseStatusLabel = "release_status"
)

var (
	chartConfigDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "chartconfig_status"),
		"Managed charts status.",
		[]string{
			chartNameLabel,
			channelNameLabel,
			releaseNameLabel,
			releaseStatusLabel,
			namespaceLabel,
		},
		nil,
	)

	cordonExpireTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "cordon_expire_time"),
		"A metric of the expire time of cordoned chartconfig as unix seconds.",
		[]string{
			chartNameLabel,
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
	c.logger.Log("level", "debug", "message", "collecting metrics for ChartConfigs")

	chartConfigs, err := c.g8sClient.CoreV1alpha1().ChartConfigs("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, chartConfig := range chartConfigs.Items {
		ch <- prometheus.MustNewConstMetric(
			chartConfigDesc,
			prometheus.GaugeValue,
			gaugeValue,
			key.ChartName(chartConfig),
			key.ChannelName(chartConfig),
			key.ReleaseName(chartConfig),
			key.ReleaseStatus(chartConfig),
			key.Namespace(chartConfig),
		)

		if key.IsCordoned(chartConfig) {
			t, err := convertToTime(key.CordonUntil(chartConfig))
			if err != nil {
				c.logger.Log("level", "warning", "message", "could not convert cordon-until", "stack", fmt.Sprintf("%#v", err))
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				cordonExpireTimeDesc,
				prometheus.GaugeValue,
				float64(t.Unix()),
				key.ChartName(chartConfig),
			)
		}
	}

	c.logger.Log("level", "debug", "message", "finished collecting metrics for ChartConfigs")

	return nil
}

// Describe emits the description for the metrics collected here.
func (a *ChartResource) Describe(ch chan<- *prometheus.Desc) error {
	ch <- chartConfigDesc
	return nil
}
