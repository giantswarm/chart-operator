package collector

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v6/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/chart-operator/v2/service/controller/chart/key"
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
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type OrphanConfigMap struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func NewOrphanConfigMap(config OrphanConfigMapConfig) (*OrphanConfigMap, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	oc := &OrphanConfigMap{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return oc, nil
}

func (oc *OrphanConfigMap) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	chartList := &v1alpha1.ChartList{}
	err := oc.k8sClient.CtrlClient().List(
		ctx,
		chartList,
	)
	if err != nil {
		return microerror.Mask(err)
	}

	desiredConfigMaps := make(map[[2]string]bool)

	for _, chart := range chartList.Items {
		desiredConfigMaps[[2]string{key.ConfigMapNamespace(chart), key.ConfigMapName(chart)}] = true
	}

	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, "app-operator"),
	}
	configMaps, err := oc.k8sClient.K8sClient().CoreV1().ConfigMaps("").List(ctx, lo)
	if err != nil {
		return microerror.Mask(err)
	}

	var orphanConfigMaps []string

	for _, cm := range configMaps.Items {
		if !desiredConfigMaps[[2]string{cm.Namespace, cm.Name}] {
			orphanConfigMaps = append(orphanConfigMaps, fmt.Sprintf("%s.%s", cm.Namespace, cm.Name))
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

	return nil
}

// Describe emits the description for the metrics collected here.
func (oc *OrphanConfigMap) Describe(ch chan<- *prometheus.Desc) error {
	ch <- orphanConfigMapDesc
	return nil
}
