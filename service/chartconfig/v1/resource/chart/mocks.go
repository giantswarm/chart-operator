package chart

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/helm"
)

type apprMock struct {
	defaultReleaseVersion string
	expectedError         bool
}

func (a *apprMock) GetReleaseVersion(customObject v1alpha1.ChartConfig) (string, error) {
	if a.expectedError {
		return "", fmt.Errorf("error getting default release")
	}

	return a.defaultReleaseVersion, nil
}

type helmMock struct {
	expectedError bool
}

func (a *helmMock) GetReleaseContent(customObject v1alpha1.ChartConfig) (*helm.ReleaseContent, error) {
	if a.expectedError {
		return nil, fmt.Errorf("error getting release content")
	}

	return &helm.ReleaseContent{}, nil
}
