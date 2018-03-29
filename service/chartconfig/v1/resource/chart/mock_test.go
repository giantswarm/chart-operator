package chart

import (
	"fmt"

	helmclient "k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/helm"
)

type apprMock struct {
	defaultReleaseVersion string
	defaultError          bool
}

func (a *apprMock) GetReleaseVersion(name, channel string) (string, error) {
	if a.defaultError {
		return "", fmt.Errorf("error getting default release")
	}

	return a.defaultReleaseVersion, nil
}

func (a *apprMock) PullChartTarball(name, channel string) (string, error) {
	return "", nil
}

type helmMock struct {
	defaultReleaseContent *helm.ReleaseContent
	defaultReleaseHistory *helm.ReleaseHistory
	defaultError          error
}

func (h *helmMock) DeleteRelease(releaseName string, options ...helmclient.DeleteOption) error {
	if h.defaultError != nil {
		return h.defaultError
	}

	return nil
}

func (h *helmMock) GetReleaseContent(releaseName string) (*helm.ReleaseContent, error) {
	if h.defaultError != nil {
		return nil, h.defaultError
	}

	return h.defaultReleaseContent, nil
}

func (h *helmMock) GetReleaseHistory(releaseName string) (*helm.ReleaseHistory, error) {
	if h.defaultError != nil {
		return nil, h.defaultError
	}

	return h.defaultReleaseHistory, nil
}

func (h *helmMock) InstallFromTarball(path, ns string, options ...helmclient.InstallOption) error {
	return nil
}
