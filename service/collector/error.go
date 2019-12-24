package collector

import (
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	chartConfigNotInstalledErrorText = "the server could not find the requested resource (get chartconfigs.core.giantswarm.io)"
)

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var chartConfigNotInstalledError = &microerror.Error{
	Kind: "chartConfigNotInstalledError",
}

// IsChartConfigNotInstalled asserts chartConfigNotInstalledError.
func IsChartConfigNotInstalled(err error) bool {
	if err == nil {
		return false
	}

	c := microerror.Cause(err)

	if strings.Contains(c.Error(), chartConfigNotInstalledErrorText) {
		return true
	}

	if c == chartConfigNotInstalledError {
		return true
	}

	return false
}
