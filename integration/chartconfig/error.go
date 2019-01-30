package chartconfig

import (
	"github.com/giantswarm/microerror"
)

var failedExecutionError = &microerror.Error{
	Kind: "failedExecutionError",
}

// IsFailedExecution asserts failedExecutionError.
func IsFailedExecution(err error) bool {
	return microerror.Cause(err) == failedExecutionError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
