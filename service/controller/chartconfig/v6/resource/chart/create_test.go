package chart

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apprclient/apprclienttest"
	"github.com/giantswarm/helmclient/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Chart_newCreate(t *testing.T) {
	testCases := []struct {
		obj               v1alpha1.ChartConfig
		currentState      *ChartState
		desiredState      *ChartState
		expectedChartName string
		description       string
	}{
		{
			description:       "empty current and desired, expected empty",
			currentState:      &ChartState{},
			desiredState:      &ChartState{},
			expectedChartName: "",
		},
		{
			description: "non-empty current, empty desired, expected empty",
			currentState: &ChartState{
				ChartName: "current",
			},
			desiredState:      &ChartState{},
			expectedChartName: "",
		},

		{
			description:  "empty current, non-empty desired, expected desired",
			currentState: &ChartState{},
			desiredState: &ChartState{
				ChartName: "desired",
			},
			expectedChartName: "desired",
		},
		{
			description: "equal non-empty current and desired, expected desired",
			currentState: &ChartState{
				ChartName: "desired",
			},
			desiredState: &ChartState{
				ChartName: "desired",
			},
			expectedChartName: "desired",
		},
		{
			description: "different non-empty current and desired, expected desired",
			currentState: &ChartState{
				ChartName: "current",
			},
			desiredState: &ChartState{
				ChartName: "desired",
			},
			expectedChartName: "desired",
		},
	}

	var newResource *Resource
	var err error
	{
		c := Config{
			ApprClient: apprclienttest.New(apprclienttest.Config{}),
			Fs:         afero.NewMemMapFs(),
			HelmClient: helmclienttest.New(helmclienttest.Config{}),
			K8sClient:  fake.NewSimpleClientset(),
			Logger:     microloggertest.New(),
		}

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newCreateChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			createChange, ok := result.(*ChartState)
			if !ok {
				t.Fatalf("expected '%T', got '%T'", createChange, result)
			}
			if createChange.ChartName != "" && createChange.ChartName != tc.expectedChartName {
				t.Fatalf("expected %s, got %s", tc.expectedChartName, createChange.ChartName)
			}
		})
	}

}
