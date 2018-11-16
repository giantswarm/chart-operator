package collector

import (
	"sync"

	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	gaugeValue       float64 = 1
	namespaceLabel           = "namespace"
	defaultNamespace         = "giantswarm"

	Namespace = "chart_operator"
)

type Config struct {
	HelmClient *helmclient.Client
	Logger     micrologger.Logger
}

type Collector struct {
	helmClient *helmclient.Client
	logger     micrologger.Logger
}

func New(config Config) (*Collector, error) {
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	c := &Collector{
		helmClient: config.HelmClient,
		logger:     config.Logger,
	}

	return c, nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- releaseDesc
	ch <- tillerReachableDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.logger.Log("level", "debug", "message", "collecting metrics")

	collectFuncs := []func(chan<- prometheus.Metric){
		c.collectChartConfigStatus,
		c.collectTillerReachable,
	}

	var wg sync.WaitGroup

	for _, collectFunc := range collectFuncs {
		wg.Add(1)

		go func(collectFunc func(ch chan<- prometheus.Metric)) {
			defer wg.Done()
			collectFunc(ch)
		}(collectFunc)
	}

	wg.Wait()

	c.logger.Log("level", "debug", "message", "finished collecting metrics")
}
