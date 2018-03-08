package chart

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/appr"
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
					},
				},
			},
			expectedState: ChartState{
				ChartName:   "quay.io/giantswarm/chart-operator-chart",
				ChannelName: "0.1-beta",
				ReleaseName: "TODO",
			},
		},
	}

	c := appr.Config{
		Address:      "http://127.0.0.1:5555",
		Logger:       microloggertest.New(),
		Organization: "giantswarm",
	}
	apprClient, err := appr.New(c)
	if err != nil {
		t.Fatalf("error == %#v, want nil", err)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				ApprClient: apprClient,
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

			if chartState != tc.expectedState {
				t.Fatalf("ChartState == %q, want %q", chartState, tc.expectedState)
			}
		})
	}

}
