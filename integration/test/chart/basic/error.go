//go:build k8srequired
// +build k8srequired

package basic

import "github.com/giantswarm/microerror"

var notDeployedError = &microerror.Error{
	Kind: "notDeployedError",
}

// IsNotDeployed asserts notDeployedError.
func IsNotDeployed(err error) bool {
	return microerror.Cause(err) == notDeployedError
}
