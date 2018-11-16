package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	chartNameLabel     = "chart_name"
	releaseStatusLabel = "release_status"
)

var (
	releaseDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "release_status"),
		"Helm release status.",
		[]string{
			chartNameLabel,
			releaseStatusLabel,
		},
		nil,
	)
)

func (c *Collector) collectChartConfigStatus(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics for releases")

	releases, err := c.helmClient.ListReleaseContents(context.Background())
	if err != nil {
		c.logger.Log("level", "error", "message", fmt.Sprintf("could not list releases"), "stack", fmt.Sprintf("%#v", err))
	}

	for _, release := range releases {
		ch <- prometheus.MustNewConstMetric(
			releaseDesc,
			prometheus.GaugeValue,
			gaugeValue,
			release.Name,
			release.Status,
		)
	}

	c.logger.Log("level", "debug", "message", "finished collecting metrics for ChartConfigs")
}
