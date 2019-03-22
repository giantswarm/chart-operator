package collector

import (
	"context"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/collector/key"
)

var (
	tillerConfiguredDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "tiller_configured"),
		"Tiller is configured correctly.",
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
		err := c.checkTillerDeployment()
		if err != nil {
			c.logger.Log("level", "error", "message", "failed to collect Tiller configuration", "stack", fmt.Sprintf("%#v", err))

			value = 0
		} else {
			value = 1
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

func (c *Collector) checkTillerDeployment() error {
	deploy, err := c.k8sClient.Extensions().Deployments(c.tillerNamespace).Get(key.TillerDeploymentName(), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	containers := deploy.Spec.Template.Spec.Containers
	if len(containers) != 1 {
		return microerror.Maskf(invalidExecutionError, "tiller container not found expected 1 got %d", len(containers))
	}

	for _, envVar := range containers[0].Env {
		if envVar.Name == key.TillerMaxHistoryEnvVarName() && envVar.Value == key.TillerMaxHistoryEnvVarValue() {
			// Configuration is correct.
			return nil
		}
	}

	return microerror.Maskf(invalidExecutionError, "tiller configuration %#q=%#q not found", key.TillerMaxHistoryEnvVarName(), key.TillerMaxHistoryEnvVarValue())
}
