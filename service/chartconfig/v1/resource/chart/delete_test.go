package chart

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Chart_newDeleteChange(t *testing.T) {
	testCases := []struct {
		obj                  v1alpha1.ChartConfig
		currentState         *ChartState
		desiredState         *ChartState
		expectedDeleteChange *ChartState
		description          string
	}{
		{
			description:          "case 0: empty current and desired, expected empty",
			currentState:         &ChartState{},
			desiredState:         &ChartState{},
			expectedDeleteChange: nil,
		},
		{
			description: "case 1: non-empty current, empty desired, expected empty",
			currentState: &ChartState{
				ChartName: "current",
			},
			desiredState:         &ChartState{},
			expectedDeleteChange: nil,
		},

		{
			description:  "case 2: empty current, non-empty desired, expected empty",
			currentState: &ChartState{},
			desiredState: &ChartState{
				ChartName:      "desired",
				ReleaseName:    "desired",
				ReleaseVersion: "desired",
			},
			expectedDeleteChange: nil,
		},
		{
			description: "case 3: equal non-empty current and desired, expected desired",
			currentState: &ChartState{
				ChartName:      "desired",
				ReleaseName:    "desired",
				ReleaseVersion: "desired",
			},
			desiredState: &ChartState{
				ChartName:      "desired",
				ReleaseName:    "desired",
				ReleaseVersion: "desired",
			},
			expectedDeleteChange: &ChartState{
				ChartName:      "desired",
				ReleaseName:    "desired",
				ReleaseVersion: "desired",
			},
		},
		{
			description: "case 4: different non-empty current and desired, expected empty",
			currentState: &ChartState{
				ChartName:      "current",
				ReleaseName:    "current",
				ReleaseVersion: "current",
			},
			desiredState: &ChartState{
				ChartName:      "desired",
				ReleaseName:    "desired",
				ReleaseVersion: "desired",
			},
			expectedDeleteChange: nil,
		},
		{
			description: "case 5: same non-empty current and desired name, different version, expected empty",
			currentState: &ChartState{
				ChartName:      "desired",
				ReleaseName:    "desired",
				ReleaseVersion: "current",
			},
			desiredState: &ChartState{
				ChartName:      "desired",
				ReleaseName:    "desired",
				ReleaseVersion: "desired",
			},
			expectedDeleteChange: nil,
		},
	}

	var newResource *Resource
	var err error
	{
		c := Config{}
		c.ApprClient = &apprMock{}
		c.HelmClient = &helmMock{}
		c.K8sClient = fake.NewSimpleClientset()
		c.Logger = microloggertest.New()

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newDeleteChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			if tc.expectedDeleteChange == nil && result != nil {
				t.Fatal("expected", nil, "got", result)
			}
			if result != nil {
				deleteChange, ok := result.(*ChartState)
				if !ok {
					t.Fatalf("expected '%T', got '%T'", deleteChange, result)
				}
				if deleteChange.ChartName != "" && deleteChange.ChartName != tc.expectedDeleteChange.ChartName {
					t.Fatalf("expected %s, got %s", tc.expectedDeleteChange, deleteChange.ChartName)
				}
			}
		})
	}

}
