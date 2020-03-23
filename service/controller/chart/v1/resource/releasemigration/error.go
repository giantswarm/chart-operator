package releasemigration

import (
	"github.com/giantswarm/microerror"
)

var releasesNotDeletedError = &microerror.Error{
	Kind: "releasesNotDeletedError",
}

// releasesNotDeletedErrorMatching asserts releasesNotDeletedError
func releasesNotDeletedErrorMatching(err error) bool {
	return microerror.Cause(err) == releasesNotDeletedError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
