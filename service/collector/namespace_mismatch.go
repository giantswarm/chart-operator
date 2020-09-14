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
	namespaceMismatch = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "namespace_mismatch"),
		"namespace is mismatching with chart CR spec.",
		[]string{
			labelChart,
			labelNamespace,
			labelPlacedNamespace,
		},
		nil,
	)
)

// NamespaceMismatchConfig is this collector's configuration struct.
type NamespaceMismatchConfig struct {
	G8sClient  versioned.Interface
	HelmClient helmclient.Interface
	Logger     micrologger.Logger
}

// NamespaceMismatch is the main struct for this collector.
type NamespaceMismatch struct {
	g8sClient  versioned.Interface
	helmClient helmclient.Interface
	logger     micrologger.Logger
}

// NewNamespaceMismatch creates a new NamespaceMismatch metrics collector.
func NewNamespaceMismatch(config NamespaceMismatchConfig) (*NamespaceMismatch, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	t := &NamespaceMismatch{
		g8sClient:  config.G8sClient,
		helmClient: config.HelmClient,
		logger:     config.Logger,
	}

	return t, nil
}

func (n *NamespaceMismatch) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	charts, err := n.g8sClient.ApplicationV1alpha1().Charts("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	for _, chart := range charts.Items {
		var value float64 = 0

		content, err := n.helmClient.GetReleaseContent(ctx, chart.Spec.Name)
		if helmclient.IsReleaseNotFound(err) {
			continue
		} else if err != nil {
			n.logger.Log("level", "warn", "message", "failed to collect namespace consistency", "stack", fmt.Sprintf("%#v", err))
			continue
		}

		if key.Namespace(chart) != content.Namespace {
			value = 1
		}

		ch <- prometheus.MustNewConstMetric(
			namespaceMismatch,
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
func (n *NamespaceMismatch) Describe(ch chan<- *prometheus.Desc) error {
	ch <- namespaceMismatch
	return nil
}
