package collector

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	tillerReachableDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "tiller_reachable"),
		"Tiller is reachable from chart-operator.",
		[]string{
			namespaceLabel,
		},
		nil,
	)
)

func (c *Collector) collectTillerReachable(ctx context.Context, ch chan<- prometheus.Metric) {
	c.logger.LogCtx(ctx, "level", "debug", "message", "collecting Tiller reachability")

	err := c.helmClient.PingTiller(ctx)
	var value float64
	if err != nil {
		c.logger.Log("level", "error", "message", "could not ping Tiller", "stack", fmt.Sprintf("%#v", err))

		value = 0
	} else {
		value = 1
	}

	ch <- prometheus.MustNewConstMetric(
		tillerReachableDesc,
		prometheus.GaugeValue,
		value,
		defaultNamespace,
	)

	c.logger.LogCtx(ctx, "level", "debug", "message", "finished collecting Tiller reachability")
}
