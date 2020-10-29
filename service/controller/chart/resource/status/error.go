package status

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var wrongStatusError = &microerror.Error{
	Kind: "wrongStatusError",
}

// IsWrongStatusError asserts wrongStatusError.
func IsWrongStatusError(err error) bool {
	return microerror.Cause(err) == wrongStatusError
}
