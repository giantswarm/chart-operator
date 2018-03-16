package helm

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	helmclient "k8s.io/helm/pkg/helm"
	helmrelease "k8s.io/helm/pkg/proto/hapi/release"
)

func Test_GetReleaseContent(t *testing.T) {
	testCases := []struct {
		description     string
		obj             v1alpha1.ChartConfig
		releases        []*helmrelease.Release
		expectedContent *ReleaseContent
		errorMatcher    func(error) bool
	}{
		{
			description: "case 0: basic match with deployed status",
			obj: v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "quay.io/giantswarm/chart-operator-chart",
						Channel: "0.1-beta",
						Release: "chart-operator",
					},
				},
			},
			releases: []*helmrelease.Release{
				helmclient.ReleaseMock(&helmclient.MockReleaseOptions{
					Name:      "chart-operator",
					Namespace: "default",
				}),
			},
			expectedContent: &ReleaseContent{
				Name:   "chart-operator",
				Status: "DEPLOYED",
				Values: map[string]interface{}{
					// Note: Values cannot be configured via the Helm mock client.
					"name": "value",
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 1: basic match with failed status",
			obj: v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "quay.io/giantswarm/chart-operator-chart",
						Channel: "0.1-beta",
						Release: "chart-operator",
					},
				},
			},
			releases: []*helmrelease.Release{
				helmclient.ReleaseMock(&helmclient.MockReleaseOptions{
					Name:       "chart-operator",
					Namespace:  "default",
					StatusCode: helmrelease.Status_FAILED,
				}),
			},
			expectedContent: &ReleaseContent{
				Name:   "chart-operator",
				Status: "FAILED",
				Values: map[string]interface{}{
					"name": "value",
				},
			},
			errorMatcher: nil,
		},
		{
			description: "case 2: chart not found",
			obj: v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "missing-chart",
						Channel: "stable",
						Release: "missing",
					},
				},
			},
			releases: []*helmrelease.Release{
				helmclient.ReleaseMock(&helmclient.MockReleaseOptions{
					Name: "chart-operator",
				}),
			},
			expectedContent: nil,
			errorMatcher:    IsReleaseNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			helm := Client{
				helmClient: &helmclient.FakeClient{
					Rels: tc.releases,
				},
				logger: microloggertest.New(),
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

			if !reflect.DeepEqual(result, tc.expectedContent) {
				t.Fatalf("Release == %q, want %q", result, tc.expectedContent)
			}
		})
	}
}
