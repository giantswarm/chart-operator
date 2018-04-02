package chart

import (
	"context"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/micrologger/microloggertest"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Chart_newUpdateChange(t *testing.T) {
	customObject := &v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name: "quay.io/giantswarm/chart-operator-chart",
			},
		},
	}
	tcs := []struct {
		description         string
		currentState        *ChartState
		desiredState        *ChartState
		expectedReleaseName string
		expectedChannelName string
	}{
		{
			description:  "empty current state, empty update change",
			currentState: &ChartState{},
			desiredState: &ChartState{
				ReleaseName: "desired-release-name",
				ChannelName: "desired-channel-name",
			},
			expectedChannelName: "",
			expectedReleaseName: "",
		},
		{
			description: "nonempty current state, different release version in desired state, expected desired state",
			currentState: &ChartState{
				ReleaseName:    "current-release-name",
				ChannelName:    "current-channel-name",
				ReleaseVersion: "current-release-version",
			},
			desiredState: &ChartState{
				ReleaseName:    "desired-release-name",
				ChannelName:    "desired-channel-name",
				ReleaseVersion: "desired-release-version",
			},
			expectedChannelName: "desired-channel-name",
			expectedReleaseName: "desired-release-name",
		},
		{
			description: "nonempty current state, equal release version in desired state, empty update change",
			currentState: &ChartState{
				ReleaseName:    "current-release-name",
				ChannelName:    "current-channel-name",
				ReleaseVersion: "release-version",
			},
			desiredState: &ChartState{
				ReleaseName:    "desired-release-name",
				ChannelName:    "desired-channel-name",
				ReleaseVersion: "release-version",
			},
			expectedChannelName: "",
			expectedReleaseName: "",
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

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newUpdateChange(context.TODO(), customObject, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			updateChange, ok := result.(*ChartState)
			if !ok {
				t.Fatalf("expected '%T', got '%T'", updateChange, result)
			}
			if updateChange.ReleaseName != tc.expectedReleaseName {
				t.Fatalf("expected ReleaseName %q, got %q", tc.expectedReleaseName, updateChange.ReleaseName)
			}
			if updateChange.ChannelName != tc.expectedChannelName {
				t.Fatalf("expected ChannelName %q, got %q", tc.expectedChannelName, updateChange.ChannelName)
			}
		})
	}
}
