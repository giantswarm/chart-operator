package collector

import (
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/chart-operator/pkg/label"
	"github.com/giantswarm/chart-operator/service/controller/chart/v1/key"
)

var (
	orphanConfigMapDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "configmap", "orphan"),
		"Configmaps without a chart CR.",
		[]string{},
		nil,
	)
)

type OrphanConfigMapConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type OrphanConfigMap struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func NewOrphanConfigMap(config OrphanConfigMapConfig) (*OrphanConfigMap, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	oc := &OrphanConfigMap{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return oc, nil
}

func (oc *OrphanConfigMap) Collect(ch chan<- prometheus.Metric) error {
	oc.logger.Log("level", "debug", "message", "collecting metrics for orphan configmaps")

	charts, err := oc.g8sClient.ApplicationV1alpha1().Charts("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	desiredConfigMaps := make(map[string]bool)

	for _, cr := range charts.Items {
		key := fmt.Sprintf("%s.%s", key.ConfigMapNamespace(cr), key.ConfigMapName(cr))
		desiredConfigMaps[key] = true
	}

	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, "app-operator"),
	}
	configMaps, err := oc.k8sClient.CoreV1().ConfigMaps("").List(lo)
	if err != nil {
		return microerror.Mask(err)
	}

	var orphanConfigMaps []string

	for _, cm := range configMaps.Items {
		key := fmt.Sprintf("%s.%s", cm.Namespace, cm.Name)

		exists, _ := desiredConfigMaps[key]
		if !exists {
			orphanConfigMaps = append(orphanConfigMaps, key)
		}
	}

	ch <- prometheus.MustNewConstMetric(
		orphanConfigMapDesc,
		prometheus.GaugeValue,
		float64(len(orphanConfigMaps)),
	)

	if len(orphanConfigMaps) > 0 {
		oc.logger.Log("level", "debug", "message", fmt.Sprintf("found %d orphan configmaps %s", len(orphanConfigMaps), strings.Join(orphanConfigMaps, " ")))
	}

	oc.logger.Log("level", "debug", "message", "finished collecting metrics for orphan configmaps")

	return nil
}

// Describe emits the description for the metrics collected here.
func (oc *OrphanConfigMap) Describe(ch chan<- *prometheus.Desc) error {
	ch <- orphanConfigMapDesc
	return nil
}
