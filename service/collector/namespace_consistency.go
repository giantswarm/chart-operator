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
)

var (
	namespaceConsistency = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "namespace_consistency"),
		"namespace is consistent with chart CR spec.",
		[]string{
			labelChart,
			labelNamespace,
			labelPlacedNamespace,
		},
		nil,
	)
)

// NamespaceConsistencyConfig is this collector's configuration struct.
type NamespaceConsistencyConfig struct {
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	Logger     micrologger.Logger
}

// NamespaceConsistency is the main struct for this collector.
type NamespaceConsistency struct {
	g8sClient  versioned.Interface
	helmClient helmclient.Interface
	logger     micrologger.Logger
}

// NewNamespaceConsistency creates a new NamespaceConsistency metrics collector.
func NewNamespaceConsistency(config NamespaceConsistencyConfig) (*NamespaceConsistency, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	t := &NamespaceConsistency{
		g8sClient:  config.G8sClient,
		helmClient: config.HelmClient,
		logger:     config.Logger,
	}

	return t, nil
}

func (n *NamespaceConsistency) Collect(ch chan<- prometheus.Metric) error {
	var value float64

	ctx := context.Background()

	charts, err := n.g8sClient.ApplicationV1alpha1().Charts("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, chart := range charts.Items {
		content, err := n.helmClient.GetReleaseContent(ctx, chart.Spec.Name)
		if err != nil {
			n.logger.Log("level", "error", "message", "failed to collect namespace consistency", "stack", fmt.Sprintf("%#v", err))
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			namespaceConsistency,
			prometheus.GaugeValue,
			value,
			content.Name,
			chart.Spec.Namespace,
			content.Namespace,
		)

	}

	return nil
}

// Describe emits the description for the metrics collected here.
func (n *NamespaceConsistency) Describe(ch chan<- *prometheus.Desc) error {
	ch <- namespaceConsistency
	return nil
}
