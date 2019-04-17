package collector

import (
	"context"
	"fmt"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
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

// TillerReachableConfig is this collector's configuration struct.
type TillerReachableConfig struct {
	HelmClient helmclient.Interface
	Helper     *helper
	Logger     micrologger.Logger

	TillerNamespace string
}

// TillerReachable is the main struct for this collector.
type TillerReachable struct {
	helmClient helmclient.Interface
	helper     *helper
	logger     micrologger.Logger

	tillerNamespace string
}

// NewTillerReachable creates a new TillerReachable metrics collector.
func NewTillerReachable(config TillerReachableConfig) (*TillerReachable, error) {
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Helper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Helper must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.TillerNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TillerNamespace must not be empty", config)
	}

	t := &TillerReachable{
		helmClient: config.HelmClient,
		helper:     config.Helper,
		logger:     config.Logger,

		tillerNamespace: config.TillerNamespace,
	}

	return t, nil
}

func (t *TillerReachable) Collect(ch chan<- prometheus.Metric) error {
	var value float64

	ctx := context.Background()

	t.logger.Log("level", "debug", "message", "collecting Tiller reachability")

	charts, err := t.helper.getCharts()
	if err != nil {
		return microerror.Mask(err)
	}

	chartConfigs, err := t.helper.getChartConfigs()
	if err != nil {
		return microerror.Mask(err)
	}

	if len(charts) == 0 && len(chartConfigs) == 0 {
		// Skip pinging tiller when there are no custom resources,
		// as tiller is only installed when there is at least one CR to reconcile.
		t.logger.Log("level", "debug", "message", "did not collect Tiller reachability")
		t.logger.Log("level", "debug", "message", "no Chart or ChartConfig CRs in the cluster")

		value = 1
	} else {
		err := t.helmClient.PingTiller(ctx)
		if err != nil {
			t.logger.Log("level", "error", "message", "failed to collect Tiller reachability", "stack", fmt.Sprintf("%#v", err))

			value = 0
		} else {
			value = 1
		}
	}

	ch <- prometheus.MustNewConstMetric(
		tillerReachableDesc,
		prometheus.GaugeValue,
		value,
		t.tillerNamespace,
	)

	t.logger.Log("level", "debug", "message", "finished collecting Tiller reachability")

	return nil
}

// Describe emits the description for the metrics collected here.
func (t *TillerReachable) Describe(ch chan<- *prometheus.Desc) error {
	ch <- tillerReachableDesc
	return nil
}
