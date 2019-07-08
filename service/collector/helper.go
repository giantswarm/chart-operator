package collector

import (
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type helperConfig struct {
	G8sClient versioned.Interface
	Logger    micrologger.Logger
}

type helper struct {
	g8sClient versioned.Interface
	logger    micrologger.Logger
}

func newHelper(config helperConfig) (*helper, error) {
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	h := &helper{
		g8sClient: config.G8sClient,
		logger:    config.Logger,
	}

	return h, nil
}

func (h *helper) getCharts() ([]*chartState, error) {
	r, err := h.g8sClient.ApplicationV1alpha1().Charts("").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	res := []*chartState{}
	for _, chart := range r.Items {
		v := &chartState{
			chartName: chart.Name,
			namespace: chart.Namespace,
		}
		res = append(res, v)
	}
	return res, nil
}

func (h *helper) getChartConfigs() ([]*chartState, error) {
	r, err := h.g8sClient.CoreV1alpha1().ChartConfigs("").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	res := []*chartState{}
	for _, chartConfig := range r.Items {
		v := &chartState{
			chartName:     chartConfig.Spec.Chart.Name,
			channelName:   chartConfig.Spec.Chart.Channel,
			namespace:     chartConfig.Namespace,
			releaseName:   chartConfig.Spec.Chart.Release,
			releaseStatus: chartConfig.Status.ReleaseStatus,
		}
		res = append(res, v)
	}
	return res, nil
}
