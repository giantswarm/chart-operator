package collector

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gaugeValue         float64 = 1
	chartNameLabel             = "chart_name"
	channelNameLabel           = "channel_name"
	releaseNameLabel           = "release_name"
	releaseStatusLabel         = "release_status"
)

type chartState struct {
	chartName     string
	channelName   string
	releaseName   string
	releaseStatus string
}

var (
	chartConfigDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "chartconfig_status"),
		"Managed charts status.",
		[]string{
			chartNameLabel,
			channelNameLabel,
			releaseNameLabel,
			releaseStatusLabel,
		},
		nil,
	)
)

func (c *Collector) collectChartConfigStatus(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics for vpcs")

	chartConfigs, err := c.getChartConfigs()
	if err != nil {
		c.logger.Log("level", "error", "message", fmt.Sprintf("could not get ChartConfigs"), "stack", fmt.Sprintf("%#v", err))
	}

	for _, chartConfig := range chartConfigs {
		ch <- prometheus.MustNewConstMetric(
			chartConfigDesc,
			prometheus.GaugeValue,
			gaugeValue,
			chartConfig.chartName,
			chartConfig.channelName,
			chartConfig.releaseName,
			chartConfig.releaseStatus,
		)
	}
	c.logger.Log("level", "debug", "message", "finished collecting metrics for ChartConfigs")
}

func (c *Collector) getChartConfigs() ([]*chartState, error) {
	r, err := c.g8sClient.CoreV1alpha1().
		ChartConfigs("giantswarm").
		List(v1.ListOptions{})

	if err != nil {
		return nil, microerror.Mask(err)
	}

	res := []*chartState{}
	for _, chartConfig := range r.Items {
		v := &chartState{
			chartName:     chartConfig.Spec.Chart.Name,
			channelName:   chartConfig.Spec.Chart.Channel,
			releaseName:   chartConfig.Spec.Chart.Release,
			releaseStatus: chartConfig.Status.ReleaseStatus,
		}
		res = append(res, v)
	}
	return res, nil
}
