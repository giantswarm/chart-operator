package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func Test_VersionBundleVersion(t *testing.T) {
	testCases := []struct {
		name            string
		obj             v1alpha1.Chart
		expectedVersion string
	}{
		{
			name: "case 0: basic match",
			obj: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"giantswarm.io/version-bundle": "1.0.0",
					},
				},
			},
			expectedVersion: "1.0.0",
		},
		{
			name: "case 1: different version",
			obj: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"giantswarm.io/version-bundle": "2.0.0",
					},
				},
			},
			expectedVersion: "2.0.0",
		},
		{
			name: "case 2: missing version",
			obj: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"test": "test",
					},
				},
			},
			expectedVersion: "",
		},
		{
			name:            "case 3: no annotations",
			obj:             v1alpha1.Chart{},
			expectedVersion: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := VersionBundleVersion(tc.obj)

			if result != tc.expectedVersion {
				t.Fatalf("VersionBundleVersion == %#q, want %#q", result, tc.expectedVersion)
			}
		})
	}
}
