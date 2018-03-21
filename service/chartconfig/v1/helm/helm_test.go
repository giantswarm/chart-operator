package helm

import (
	"reflect"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	helmclient "k8s.io/helm/pkg/helm"
	helmchart "k8s.io/helm/pkg/proto/hapi/chart"
	helmrelease "k8s.io/helm/pkg/proto/hapi/release"
)

func Test_GetReleaseContent(t *testing.T) {
	testCases := []struct {
		description     string
		releaseName     string
		releases        []*helmrelease.Release
		expectedContent *ReleaseContent
		errorMatcher    func(error) bool
	}{
		{
			description: "case 0: basic match with deployed status",
			releaseName: "chart-operator",
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
			releaseName: "chart-operator",
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
			releaseName: "missing",
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
			result, err := helm.GetReleaseContent(tc.releaseName)

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

func Test_GetReleaseHistory(t *testing.T) {
	testCases := []struct {
		description     string
		releaseName     string
		releases        []*helmrelease.Release
		expectedHistory *ReleaseHistory
		errorMatcher    func(error) bool
	}{
		{
			description: "case 0: basic match with version",
			releaseName: "chart-operator",
			releases: []*helmrelease.Release{
				helmclient.ReleaseMock(&helmclient.MockReleaseOptions{
					Name:      "chart-operator",
					Namespace: "default",
					Chart: &helmchart.Chart{
						Metadata: &helmchart.Metadata{
							Version: "0.1.0",
						},
					},
				}),
			},
			expectedHistory: &ReleaseHistory{
				Name:    "chart-operator",
				Version: "0.1.0",
			},
			errorMatcher: nil,
		},
		{
			description: "case 1: different version",
			releaseName: "chart-operator",
			releases: []*helmrelease.Release{
				helmclient.ReleaseMock(&helmclient.MockReleaseOptions{
					Name:      "chart-operator",
					Namespace: "default",
					Chart: &helmchart.Chart{
						Metadata: &helmchart.Metadata{
							Version: "1.0.0-rc1",
						},
					},
				}),
			},
			expectedHistory: &ReleaseHistory{
				Name:    "chart-operator",
				Version: "1.0.0-rc1",
			},
			errorMatcher: nil,
		},
		{
			description: "case 2: too many results",
			releaseName: "missing",
			releases: []*helmrelease.Release{
				helmclient.ReleaseMock(&helmclient.MockReleaseOptions{
					Name:      "chart-operator",
					Namespace: "default",
					Chart: &helmchart.Chart{
						Metadata: &helmchart.Metadata{
							Version: "1.0.0-rc1",
						},
					},
				}),
				helmclient.ReleaseMock(&helmclient.MockReleaseOptions{
					Name:      "chart-operator",
					Namespace: "default",
					Chart: &helmchart.Chart{
						Metadata: &helmchart.Metadata{
							Version: "1.0.0-rc1",
						},
					},
				}),
			},
			expectedHistory: nil,
			errorMatcher:    IsTooManyResults,
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
			result, err := helm.GetReleaseHistory(tc.releaseName)

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

			if !reflect.DeepEqual(result, tc.expectedHistory) {
				t.Fatalf("Release == %q, want %q", result, tc.expectedHistory)
			}
		})
	}
}
