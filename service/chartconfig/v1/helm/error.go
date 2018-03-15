package helm

import (
	"strings"

	"github.com/giantswarm/microerror"
)

const (
	releaseNotFoundErrorPrefix = "No such release:"
)

var invalidConfigError = microerror.New("invalid config")

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var releaseNotFoundError = microerror.New("release not found")

// IsReleaseNotFound asserts releaseNotFoundError.
func IsReleaseNotFound(err error) bool {
	if strings.HasPrefix(err.Error(), releaseNotFoundErrorPrefix) {
		return true
	}
	if microerror.Cause(err) == releaseNotFoundError {
		return true
	}

	return false
}
