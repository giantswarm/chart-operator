package collector

import (
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chartconfig/v7/key"
)

var (
	chartConfigDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "chartconfig", "status"),
		"Managed charts status.",
		[]string{
			labelChart,
			labelChannel,
			labelRelease,
			labelReleaseStatus,
			labelNamespace,
		},
		nil,
	)

	chartConfigCordonExpireTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "chartconfig", "cordon_expire_time_seconds"),
		"A metric of the expire time of cordoned chartconfig as unix seconds.",
		[]string{
			labelChart,
		},
		nil,
	)
)

// ChartConfigResourceConfig is this collector's configuration struct.
type ChartConfigResourceConfig struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

// ChartConfigResource is the main struct for this collector.
type ChartConfigResource struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

// NewChartConfigResource creates a new ChartConfigResource metrics collector.
func NewChartConfigResource(config ChartConfigResourceConfig) (*ChartConfigResource, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	a := &ChartConfigResource{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return a, nil
}

func (c *ChartConfigResource) Collect(ch chan<- prometheus.Metric) error {
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

		if !key.IsCordoned(chartConfig) {
			continue
		}

		t, err := convertToTime(key.CordonUntil(chartConfig))
		if err != nil {
			c.logger.Log("level", "warning", "message", "could not convert cordon-until", "stack", fmt.Sprintf("%#v", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			chartConfigCordonExpireTimeDesc,
			prometheus.GaugeValue,
			float64(t.Unix()),
			key.ChartName(chartConfig),
		)
	}

	c.logger.Log("level", "debug", "message", "finished collecting metrics for ChartConfigs")

	return nil
}

// Describe emits the description for the metrics collected here.
func (a *ChartConfigResource) Describe(ch chan<- *prometheus.Desc) error {
	ch <- chartConfigDesc
	ch <- chartConfigCordonExpireTimeDesc
	return nil
}

func convertToTime(datetime string) (time.Time, error) {
	layout := "2006-01-02T15:04:05"

	split := strings.Split(datetime, ".")
	if len(split) == 0 {
		return time.Time{}, microerror.Maskf(invalidExecutionError, "'%#v' must have at least one item in order to collect metrics for the cordon expiration", datetime)
	}

	t, err := time.Parse(layout, split[0])

	if err != nil {
		return time.Time{}, microerror.Mask(err)
	}

	return t, nil
}
