package key

import (
	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/microerror"
)

// ToCustomObject converts value to v1alpha1.ChartConfig and returns it or error
// if type does not match.
func ToCustomObject(v interface{}) (v1alpha1.ChartConfig, error) {
	customObjectPointer, ok := v.(*v1alpha1.ChartConfig)
	if !ok {
		return v1alpha1.ChartConfig{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &v1alpha1.ChartConfig{}, v)
	}

	if customObjectPointer == nil {
		return v1alpha1.ChartConfig{}, microerror.Maskf(emptyValueError,
			"empty value cannot be converted to CustomObject")
	}

	return *customObjectPointer, nil
}
