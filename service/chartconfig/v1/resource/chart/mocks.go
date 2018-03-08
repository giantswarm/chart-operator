package chart

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
)

type apprMock struct {
	defaultRelease string
	expectedError  bool
}

func (a *apprMock) GetRelease(customObject v1alpha1.ChartConfig) (string, error) {
	if a.expectedError {
		return "", fmt.Errorf("error getting default release")
	}

	return a.defaultRelease, nil
}
