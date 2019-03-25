package collector

import (
	"context"
	"fmt"
	"strconv"

	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/collector/key"
)

var (
	tillerConfiguredDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "tiller_max_history"),
		"Tiller setting for number of revisions to save per release.",
		[]string{
			namespaceLabel,
		},
		nil,
	)
)

func (c *Collector) collectTillerConfigured(ctx context.Context, ch chan<- prometheus.Metric) {
	var value float64

	c.logger.LogCtx(ctx, "level", "debug", "message", "collecting Tiller configuration")

	charts, err := c.getCharts()
	if err != nil {
		c.logger.LogCtx(ctx, "level", "debug", "message", "could not get Charts", "stack", fmt.Sprintf("%#v", err))
		return
	}

	chartConfigs, err := c.getChartConfigs()
	if err != nil {
		c.logger.LogCtx(ctx, "level", "debug", "message", "could not get ChartConfigs", "stack", fmt.Sprintf("%#v", err))
		return
	}

	if len(charts) == 0 && len(chartConfigs) == 0 {
		// Skip checking tiller when there are no custom resources,
		// as tiller is only installed when there is at least one CR to reconcile.
		c.logger.Log("level", "message", "message", "did not collect Tiller configuration")
		c.logger.Log("level", "message", "message", "no Chart or ChartConfig CRs in the cluster")

		value = 1
	} else {
		value, err = c.getTillerMaxHistory()
		if err != nil {
			c.logger.Log("level", "error", "message", "failed to get Tiller max history", "stack", fmt.Sprintf("%#v", err))
		}
	}

	ch <- prometheus.MustNewConstMetric(
		tillerConfiguredDesc,
		prometheus.GaugeValue,
		value,
		c.tillerNamespace,
	)

	c.logger.LogCtx(ctx, "level", "debug", "message", "finished collecting Tiller configuration")
}

func (c *Collector) getTillerMaxHistory() (float64, error) {
	deploy, err := c.k8sClient.Extensions().Deployments(c.tillerNamespace).Get(key.TillerDeploymentName(), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return 0, nil
	} else if err != nil {
		return 0, microerror.Mask(err)
	}

	containers := deploy.Spec.Template.Spec.Containers
	if len(containers) != 1 {
		return 0, microerror.Maskf(invalidExecutionError, "tiller container not found expected 1 got %d", len(containers))
	}

	for _, envVar := range containers[0].Env {
		if envVar.Name == key.TillerMaxHistoryEnvVarName() {
			value, err := strconv.ParseFloat(envVar.Value, 64)
			if err != nil {
				return 0, microerror.Mask(err)
			}

			return value, nil
		}
	}

	return 0, microerror.Maskf(invalidExecutionError, "tiller env var %#q not found", key.TillerMaxHistoryEnvVarName())
}
