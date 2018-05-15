package basic

import (
	"github.com/giantswarm/microerror"
)

var releaseStatusNotMatchingError = microerror.New("release status not matching")

// IsReleaseStatusNotMatching asserts releaseStatusNotMatchingError
func IsReleaseStatusNotMatching(err error) bool {
	return microerror.Cause(err) == releaseStatusNotMatchingError
}

var releaseVersionNotMatchingError = microerror.New("release version not matching")

// IsReleaseVesionNotMatching asserts releaseVersionNotMatchingError
func IsReleaseVesionNotMatching(err error) bool {
	return microerror.Cause(err) == releaseVersionNotMatchingError
}
