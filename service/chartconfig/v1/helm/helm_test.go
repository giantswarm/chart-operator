package helm

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	helmclient "k8s.io/helm/pkg/helm"
)

func Test_GetReleaseContent(t *testing.T) {
	testCases := []struct {
		description     string
		obj             v1alpha1.ChartConfig
		expectedRelease *Release
		errorMatcher    func(error) bool
	}{
		{
			description: "case 0: chart not found",
			obj: v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "missing-chart",
						Channel: "stable",
						Release: "missing",
					},
				},
			},
			expectedRelease: nil,
			errorMatcher:    IsNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			helm := Client{
				helmClient: &helmclient.FakeClient{},
				logger:     microloggertest.New(),
			}

			result, err := helm.GetReleaseContent(tc.obj)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// correct; carry on
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(result, tc.expectedRelease) {
				t.Fatalf("Release == %q, want %q", result, tc.expectedRelease)
			}
		})
	}
}
