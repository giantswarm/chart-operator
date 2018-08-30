package collector

import (
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
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
	G8sClient  versioned.Interface
	HelmClient *helmclient.Client
	Logger     micrologger.Logger
}

type Collector struct {
	g8sClient  versioned.Interface
	helmClient *helmclient.Client
	logger     micrologger.Logger
}

func New(config Config) (*Collector, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	c := &Collector{
		g8sClient:  config.G8sClient,
		helmClient: config.HelmClient,
		logger:     config.Logger,
	}

	return c, nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- chartConfigDesc
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
