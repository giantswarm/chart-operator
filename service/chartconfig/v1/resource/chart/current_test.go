package chart

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/helm"
)

func Test_CurrentState(t *testing.T) {
	testCases := []struct {
		name           string
		obj            *v1alpha1.ChartConfig
		releaseContent *helm.ReleaseContent
		releaseHistory *helm.ReleaseHistory
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
			releaseContent: &helm.ReleaseContent{
				Name:   "chart-operator",
				Status: "DEPLOYED",
				Values: map[string]interface{}{
					"key": "value",
				},
			},
			releaseHistory: &helm.ReleaseHistory{
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
				ReleaseStatus:  "DEPLOYED",
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
			releaseContent: &helm.ReleaseContent{
				Name:   "chart-operator",
				Status: "FAILED",
				Values: map[string]interface{}{
					"foo": "bar",
				},
			},
			releaseHistory: &helm.ReleaseHistory{
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
				ReleaseStatus:  "FAILED",
				ReleaseVersion: "0.1.3",
			},
		},
		{
			name: "case 2: error expected",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "quay.io/giantswarm/chart-operator-chart",
						Channel: "0.1-beta",
						Release: "missing-operator",
					},
				},
			},
			releaseContent: &helm.ReleaseContent{},
			releaseHistory: &helm.ReleaseHistory{},
			expectedError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			helmClient := &helmMock{
				expectedError:         tc.expectedError,
				defaultReleaseContent: tc.releaseContent,
				defaultReleaseHistory: tc.releaseHistory,
			}

			c := Config{
				ApprClient: &apprMock{},
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
			case err != nil && tc.expectedError == false:
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
