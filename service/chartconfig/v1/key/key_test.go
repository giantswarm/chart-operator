package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
)

func Test_ChartName(t *testing.T) {
	expectedName := "chart-operator-chart"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
			},
		},
	}

	if ChartName(obj) != expectedName {
		t.Fatalf("chart name %s, want %s", ChartName(obj), expectedName)
	}
}

func Test_ChannelName(t *testing.T) {
	expectedChannel := "0.1-beta"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
			},
		},
	}

	if ChannelName(obj) != expectedChannel {
		t.Fatalf("chart name %s, want %s", ChannelName(obj), expectedChannel)
	}
}

func Test_ToCustomObject(t *testing.T) {
	testCases := []struct {
		name           string
		input          interface{}
		expectedObject v1alpha1.ChartConfig
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: basic match",
			input: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "chart-operator-chart",
						Channel: "0.1-beta",
					},
				},
			},
			expectedObject: v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "chart-operator-chart",
						Channel: "0.1-beta",
					},
				},
			},
		},
		{
			name:         "case 1: wrong type",
			input:        &v1alpha1.CertConfig{},
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
				t.Fatalf("Custom Object == %q, want %q", result, tc.expectedObject)
			}
		})
	}
}
