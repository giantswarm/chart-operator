package chart

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/helm"
)

type apprMock struct {
	defaultReleaseVersion string
	defaultError          bool
}

func (a *apprMock) GetReleaseVersion(customObject v1alpha1.ChartConfig) (string, error) {
	if a.defaultError {
		return "", fmt.Errorf("error getting default release")
	}

	return a.defaultReleaseVersion, nil
}

func (a *apprMock) PullChartTarball(v1alpha1.ChartConfig) (string, error) {
	return "", nil
}

type helmMock struct {
	defaultReleaseContent *helm.ReleaseContent
	defaultReleaseHistory *helm.ReleaseHistory
	defaultError          error
}

func (h *helmMock) GetReleaseContent(customObject v1alpha1.ChartConfig) (*helm.ReleaseContent, error) {
	if h.defaultError != nil {
		return nil, h.defaultError
	}

	return h.defaultReleaseContent, nil
}

func (h *helmMock) GetReleaseHistory(customObject v1alpha1.ChartConfig) (*helm.ReleaseHistory, error) {
	if h.defaultError != nil {
		return nil, h.defaultError
	}

	return h.defaultReleaseHistory, nil
}

func (h *helmMock) InstallFromTarball(path string) error {
	return nil
}
