package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
)

func Test_ToCustomObject(t *testing.T) {
	testCases := []struct {
		name           string
		input          interface{}
		expectedObject v1alpha1.Chart
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: basic match",
			input: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name:       "prometheus-chart",
					Namespace:  "monitoring",
					TarballURL: "https://path.to/chart.tgz",
				},
			},
			expectedObject: v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name:       "prometheus-chart",
					Namespace:  "monitoring",
					TarballURL: "https://path.to/chart.tgz",
				},
			},
		},
		{
			name:         "case 1: wrong type",
			input:        &v1alpha1.App{},
			errorMatcher: IsWrongTypeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ToCustomObject(tc.input)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(result, tc.expectedObject) {
				t.Fatalf("Custom Object == %#v, want %#v", result, tc.expectedObject)
			}
		})
	}
}
