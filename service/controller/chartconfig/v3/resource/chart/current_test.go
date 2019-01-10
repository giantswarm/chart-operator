package chart

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apprclient/apprclienttest"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/helmclient/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_CurrentState(t *testing.T) {
	testCases := []struct {
		name           string
		obj            *v1alpha1.ChartConfig
		releaseContent *helmclient.ReleaseContent
		releaseHistory *helmclient.ReleaseHistory
		returnedError  error
		expectedState  ChartState
		expectedError  bool
	}{
		{
			name: "case 0: basic match",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "quay.io/giantswarm/chart-operator-chart",
						Channel: "0.1-beta",
						Release: "chart-operator",
					},
				},
			},
			releaseContent: &helmclient.ReleaseContent{
				Name:   "chart-operator",
				Status: "DEPLOYED",
				Values: map[string]interface{}{
					"key": "value",
				},
			},
			releaseHistory: &helmclient.ReleaseHistory{
				Name:    "chart-operator",
				Version: "0.1.2",
			},
			expectedState: ChartState{
				ChartName: "quay.io/giantswarm/chart-operator-chart",
				ChartValues: map[string]interface{}{
					"key": "value",
				},
				ChannelName:    "0.1-beta",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.2",
			},
		},
		{
			name: "case 1: different values",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "quay.io/giantswarm/chart-operator-chart",
						Channel: "0.1-beta",
						Release: "chart-operator",
					},
				},
			},
			releaseContent: &helmclient.ReleaseContent{
				Name:   "chart-operator",
				Status: "FAILED",
				Values: map[string]interface{}{
					"foo": "bar",
				},
			},
			releaseHistory: &helmclient.ReleaseHistory{
				Name:    "chart-operator",
				Version: "0.1.3",
			},
			expectedState: ChartState{
				ChartName: "quay.io/giantswarm/chart-operator-chart",
				ChartValues: map[string]interface{}{
					"foo": "bar",
				},
				ChannelName:    "0.1-beta",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.3",
			},
		},
		{
			name: "case 2: empty state when error for no release present",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "quay.io/giantswarm/chart-operator-chart",
						Channel: "0.1-beta",
						Release: "missing-operator",
					},
				},
			},
			releaseContent: &helmclient.ReleaseContent{},
			releaseHistory: &helmclient.ReleaseHistory{},
			returnedError:  fmt.Errorf("No such release: missing-operator"),
		},
		{
			name: "case 3: unexpected error",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "quay.io/giantswarm/chart-operator-chart",
						Channel: "0.1-beta",
						Release: "missing-operator",
					},
				},
			},
			releaseContent: &helmclient.ReleaseContent{},
			releaseHistory: &helmclient.ReleaseHistory{},
			returnedError:  fmt.Errorf("Unexpected error"),
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var helmClient helmclient.Interface
			{
				c := helmclienttest.Config{
					DefaultReleaseContent: tc.releaseContent,
					DefaultReleaseHistory: tc.releaseHistory,
					DefaultError:          tc.returnedError,
				}
				helmClient = helmclienttest.New(c)
			}

			c := Config{
				ApprClient: apprclienttest.New(apprclienttest.Config{}),
				Fs:         afero.NewMemMapFs(),
				HelmClient: helmClient,
				K8sClient:  fake.NewSimpleClientset(),
				Logger:     microloggertest.New(),
			}

			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := r.GetCurrentState(context.TODO(), tc.obj)
			switch {
			case err != nil && !tc.expectedError:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.expectedError:
				t.Fatalf("error == nil, want non-nil")
			}

			chartState, err := toChartState(result)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			if !reflect.DeepEqual(chartState, tc.expectedState) {
				t.Fatalf("ChartState == %q, want %q", chartState, tc.expectedState)
			}
		})
	}

}
