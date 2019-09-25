package collector

import (
	"fmt"
	"strconv"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/service/collector/key"
)

var (
	tillerConfiguredDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "", "tiller_max_history"),
		"Tiller setting for number of revisions to save per release.",
		[]string{
			labelNamespace,
		},
		nil,
	)
)

// TillerMaxHistoryConfig is this collector's configuration struct.
type TillerMaxHistoryConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	TillerNamespace string
}

// TillerMaxHistory is the main struct for this collector.
type TillerMaxHistory struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	tillerNamespace string
}

// NewTillerMaxHistory creates a new TillerMaxHistory metrics collector.
func NewTillerMaxHistory(config TillerMaxHistoryConfig) (*TillerMaxHistory, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.TillerNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TillerNamespace must not be empty", config)
	}

	t := &TillerMaxHistory{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		tillerNamespace: config.TillerNamespace,
	}

	return t, nil
}

func (t *TillerMaxHistory) Collect(ch chan<- prometheus.Metric) error {
	var value float64

	t.logger.Log("level", "debug", "message", "collecting Tiller max history")

	charts, err := t.g8sClient.ApplicationV1alpha1().Charts("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	chartConfigs, err := t.g8sClient.CoreV1alpha1().ChartConfigs("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	if len(charts.Items) == 0 && len(chartConfigs.Items) == 0 {
		// Skip checking tiller when there are no custom resources,
		// as tiller is only installed when there is at least one CR to reconcile.
		t.logger.Log("level", "debug", "message", "did not collect Tiller max history")
		t.logger.Log("level", "debug", "message", "no Chart or ChartConfig CRs in the cluster")

		value = 1
	} else {
		value, err = t.getTillerMaxHistory()
		if err != nil {
			t.logger.Log("level", "error", "message", "failed to get Tiller max history", "stack", fmt.Sprintf("%#v", err))
		}
	}

	ch <- prometheus.MustNewConstMetric(
		tillerConfiguredDesc,
		prometheus.GaugeValue,
		value,
		t.tillerNamespace,
	)

	t.logger.Log("level", "debug", "message", "finished collecting Tiller max history")

	return nil
}

// Describe emits the description for the metrics collected here.
func (t *TillerMaxHistory) Describe(ch chan<- *prometheus.Desc) error {
	ch <- tillerConfiguredDesc
	return nil
}

func (t *TillerMaxHistory) getTillerMaxHistory() (float64, error) {
	deploy, err := t.k8sClient.ExtensionsV1beta1().Deployments(t.tillerNamespace).Get(key.TillerDeploymentName(), metav1.GetOptions{})
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
