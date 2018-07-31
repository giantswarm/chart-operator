// +build k8srequired

package release

import "github.com/giantswarm/microerror"

var releaseStatusNotMatchingError = &microerror.Error{
	Kind: "releaseStatusNotMatchingError",
}

// IsReleaseStatusNotMatching asserts releaseStatusNotMatchingError
func IsReleaseStatusNotMatching(err error) bool {
	return microerror.Cause(err) == releaseStatusNotMatchingError
}

var releaseVersionNotMatchingError = &microerror.Error{
	Kind: "releaseVersionNotMatchingError",
}

// IsReleaseVersionNotMatching asserts releaseVersionNotMatchingError
func IsReleaseVersionNotMatching(err error) bool {
	return microerror.Cause(err) == releaseVersionNotMatchingError
}
