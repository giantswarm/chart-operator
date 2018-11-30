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
	var value float64

	c.logger.LogCtx(ctx, "level", "debug", "message", "collecting Tiller reachability")
	chartConfigs, err := c.getChartConfigs()
	if err != nil {
		c.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("could not get ChartConfigs"), "stack", fmt.Sprintf("%#v", err))
	}
	if err == nil && len(chartConfigs) == 0 {
		// Skip pinging tiller when there is no ChartConfig,
		// as tiller is only installed when there is at least one ChartConfig to reconcile.
		c.logger.Log("level", "error", "message", "did not collect Tiller reachability")
		c.logger.Log("level", "error", "message", "no ChartConfg CRs in the cluster")

		value = 1
	} else {
		err := c.helmClient.PingTiller(ctx)
		if err != nil {
			c.logger.Log("level", "error", "message", "failed to collect Tiller reachability", "stack", fmt.Sprintf("%#v", err))

			value = 0
		} else {
			value = 1
		}
	}

	ch <- prometheus.MustNewConstMetric(
		tillerReachableDesc,
		prometheus.GaugeValue,
		value,
		defaultNamespace,
	)

	c.logger.LogCtx(ctx, "level", "debug", "message", "finished collecting Tiller reachability")
}
