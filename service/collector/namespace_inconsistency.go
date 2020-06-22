package collector

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/service/controller/chart/key"
)

var (
	namespaceInconsistency = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "namespace_inconsistency"),
		"namespace is consistent with chart CR spec.",
		[]string{
			labelChart,
			labelNamespace,
			labelPlacedNamespace,
		},
		nil,
	)
)

// NamespaceInconsistencyConfig is this collector's configuration struct.
type NamespaceInconsistencyConfig struct {
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	Logger     micrologger.Logger
}

// NamespaceInconsistency is the main struct for this collector.
type NamespaceInconsistency struct {
	g8sClient  versioned.Interface
	helmClient helmclient.Interface
	logger     micrologger.Logger
}

// NewNamespaceInconsistency creates a new NamespaceInconsistency metrics collector.
func NewNamespaceInconsistency(config NamespaceInconsistencyConfig) (*NamespaceInconsistency, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	t := &NamespaceInconsistency{
		g8sClient:  config.G8sClient,
		helmClient: config.HelmClient,
		logger:     config.Logger,
	}

	return t, nil
}

func (n *NamespaceInconsistency) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	charts, err := n.g8sClient.ApplicationV1alpha1().Charts("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	var value float64
	for _, chart := range charts.Items {
		content, err := n.helmClient.GetReleaseContent(ctx, chart.Spec.Name)
		if err != nil {
			n.logger.Log("level", "error", "message", "failed to collect namespace consistency", "stack", fmt.Sprintf("%#v", err))
			continue
		}

		if key.Namespace(chart) != content.Namespace {
			value = 1
		}

		ch <- prometheus.MustNewConstMetric(
			namespaceInconsistency,
			prometheus.GaugeValue,
			value,
			content.Name,
			key.Namespace(chart),
			content.Namespace,
		)

	}

	return nil
}

// Describe emits the description for the metrics collected here.
func (n *NamespaceInconsistency) Describe(ch chan<- *prometheus.Desc) error {
	ch <- namespaceInconsistency
	return nil
}
