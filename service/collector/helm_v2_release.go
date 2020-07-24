package collector

import (
	"context"
	"fmt"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	helmV2ReleaseDesc *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(Namespace, "helm", "release"),
		"Dangling Helm V2 releases",
		[]string{},
		nil,
	)
)

type HelmV2ReleaseConfig struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	TillerNamespace string
}

type HelmV2Release struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	tillerNamespace string
}

func NewHelmV2Release(config HelmV2ReleaseConfig) (*HelmV2Release, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.TillerNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.TillerNamespace must not be empty", config)
	}

	h := &HelmV2Release{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		tillerNamespace: config.TillerNamespace,
	}

	return h, nil
}

func (h *HelmV2Release) Collect(ch chan<- prometheus.Metric) error {
	lo := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", "OWNER", "TILLER"),
	}

	// Check whether helm 2 release configMaps still exist.
	cms, err := h.k8sClient.CoreV1().ConfigMaps(h.tillerNamespace).List(context.Background(), lo)
	if err != nil {
		return microerror.Mask(err)
	}

	hasReleases := map[string]bool{}
	for _, cm := range cms.Items {
		name := cm.GetLabels()["NAME"]
		if _, ok := hasReleases[name]; !ok {
			hasReleases[name] = true
		}
	}

	releases := make([]string, 0, len(hasReleases))
	for k := range hasReleases {
		releases = append(releases, k)
	}

	ch <- prometheus.MustNewConstMetric(
		helmV2ReleaseDesc,
		prometheus.GaugeValue,
		float64(len(releases)),
	)

	if len(releases) > 0 {
		h.logger.Log("level", "debug", "message", fmt.Sprintf("found %d helm v2 releases; %s", len(releases), strings.Join(releases, " ")))
	}

	return nil
}

// Describe emits the description for the metrics collected here.
func (h *HelmV2Release) Describe(ch chan<- *prometheus.Desc) error {
	ch <- helmV2ReleaseDesc
	return nil
}
