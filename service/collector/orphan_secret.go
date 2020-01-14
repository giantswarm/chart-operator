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
	orphanSecretDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "secret", "orphan"),
		"Secrets without a chart CR.",
		[]string{},
		nil,
	)
)

type OrphanSecretConfig struct {
	G8sClient versioned.Interface
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type OrphanSecret struct {
	g8sClient versioned.Interface
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func NewOrphanSecret(config OrphanSecretConfig) (*OrphanSecret, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	oc := &OrphanSecret{
		g8sClient: config.G8sClient,
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return oc, nil
}

func (oc *OrphanSecret) Collect(ch chan<- prometheus.Metric) error {
	charts, err := oc.g8sClient.ApplicationV1alpha1().Charts("").List(metav1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	desiredSecrets := make(map[[2]string]bool)

	for _, chart := range charts.Items {
		desiredSecrets[[2]string{key.SecretNamespace(chart), key.SecretName(chart)}] = true
	}

	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", label.ManagedBy, "app-operator"),
	}
	secrets, err := oc.k8sClient.CoreV1().Secrets("").List(lo)
	if err != nil {
		return microerror.Mask(err)
	}

	var orphanSecrets []string

	for _, cm := range secrets.Items {
		if !desiredSecrets[[2]string{cm.Namespace, cm.Name}] {
			orphanSecrets = append(orphanSecrets, fmt.Sprintf("%s.%s", cm.Namespace, cm.Name))
		}
	}

	ch <- prometheus.MustNewConstMetric(
		orphanSecretDesc,
		prometheus.GaugeValue,
		float64(len(orphanSecrets)),
	)

	if len(orphanSecrets) > 0 {
		oc.logger.Log("level", "debug", "message", fmt.Sprintf("found %d orphan secrets %s", len(orphanSecrets), strings.Join(orphanSecrets, " ")))
	}

	return nil
}

// Describe emits the description for the metrics collected here.
func (oc *OrphanSecret) Describe(ch chan<- *prometheus.Desc) error {
	ch <- orphanSecretDesc
	return nil
}
