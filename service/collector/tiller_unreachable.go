package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	tillerUnreachableDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "tiller_unreachable"),
		"Tiller is not reachable from chart-operator.",
		[]string{},
		nil,
	)
)

func (c *Collector) collectTillerUnreachable(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting Tiller reachability")

	err := c.helmClient.PingTiller()
	if err != nil {
		c.logger.Log("level", "error", "message", "could not ping Tiller", "stack", fmt.Sprintf("%#v", err))

		ch <- prometheus.MustNewConstMetric(
			tillerUnreachableDesc,
			prometheus.GaugeValue,
			gaugeValue,
		)
	}

	c.logger.Log("level", "debug", "message", "finished collecting Tiller reachability")
}
