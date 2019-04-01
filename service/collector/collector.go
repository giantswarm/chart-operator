package collector

import (
	"context"
	"sync"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
)

const (
	gaugeValue     float64 = 1
	namespaceLabel         = "namespace"

	Namespace = "chart_operator"
)

type Config struct {
	G8sClient  versioned.Interface
	K8sClient  kubernetes.Interface
	HelmClient *helmclient.Client
	Logger     micrologger.Logger

	TillerNamespace string
	WatchNamespace  string
}

type Collector struct {
	g8sClient  versioned.Interface
	k8sClient  kubernetes.Interface
	helmClient *helmclient.Client
	logger     micrologger.Logger

	tillerNamespace string
	watchNamespace  string
}

func New(config Config) (*Collector, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.TillerNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TillerNamespace must not be empty", config)
	}

	c := &Collector{
		g8sClient:  config.G8sClient,
		k8sClient:  config.K8sClient,
		helmClient: config.HelmClient,
		logger:     config.Logger,

		tillerNamespace: config.TillerNamespace,
		watchNamespace:  config.WatchNamespace,
	}

	return c, nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- chartConfigDesc
	ch <- tillerReachableDesc
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	ctx := context.Background()

	c.logger.LogCtx(ctx, "level", "debug", "message", "collecting metrics")

	collectFuncs := []func(context.Context, chan<- prometheus.Metric){
		c.collectChartConfigStatus,
		c.collectTillerConfigured,
		c.collectTillerReachable,
	}

	var wg sync.WaitGroup

	for _, collectFunc := range collectFuncs {
		wg.Add(1)

		go func(collectFunc func(ctx context.Context, ch chan<- prometheus.Metric)) {
			defer wg.Done()
			collectFunc(ctx, ch)
		}(collectFunc)
	}

	wg.Wait()

	c.logger.LogCtx(ctx, "level", "debug", "message", "finished collecting metrics")
}
