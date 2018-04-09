package chart

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_DesiredState(t *testing.T) {
	testCases := []struct {
		name          string
		obj           *v1alpha1.ChartConfig
		expectedState ChartState
		errorMatcher  func(error) bool
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
			expectedState: ChartState{
				ChartName:      "quay.io/giantswarm/chart-operator-chart",
				ChannelName:    "0.1-beta",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.2",
			},
		},
	}

	apprClient := &apprMock{
		defaultReleaseVersion: "0.1.2",
	}
	helmClient := &helmMock{
		defaultError: nil,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				ApprClient: apprClient,
				Fs:         afero.NewMemMapFs(),
				HelmClient: helmClient,
				K8sClient:  fake.NewSimpleClientset(),
				Logger:     microloggertest.New(),
			}

			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := r.GetDesiredState(context.TODO(), tc.obj)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
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
