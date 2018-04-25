package key

import "github.com/giantswarm/microerror"

var emptyValueError = microerror.New("empty value")

// IsEmptyValueError asserts emptyValueError.
func IsEmptyValueError(err error) bool {
	return microerror.Cause(err) == emptyValueError
}

var wrongTypeError = microerror.New("wrong type")

// IsWrongTypeError asserts wrongTypeError.
func IsWrongTypeError(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}
