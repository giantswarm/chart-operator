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
		expectedUpdateState *ChartState
	}{
		{
			description:  "case 0: empty current state, empty update change",
			currentState: &ChartState{},
			desiredState: &ChartState{
				ReleaseName: "desired-release-name",
				ChannelName: "desired-channel-name",
			},
			expectedUpdateState: nil,
		},
		{
			description: "case 1: nonempty current state, equal desired state, empty update change",
			currentState: &ChartState{
				ReleaseName:    "release-name",
				ChannelName:    "channel-name",
				ReleaseVersion: "release-version",
			},
			desiredState: &ChartState{
				ReleaseName:    "release-name",
				ChannelName:    "channel-name",
				ReleaseVersion: "release-version",
			},
			expectedUpdateState: nil,
		},
		{
			description: "case 2: nonempty current state, different release version in desired state, expected desired state",
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
			expectedUpdateState: &ChartState{
				ChannelName: "desired-channel-name",
				ReleaseName: "desired-release-name",
			},
		},
		{
			description: "case 3: nonempty current state, desired state has values, expected desired state",
			currentState: &ChartState{
				ReleaseName:    "release-name",
				ChannelName:    "channel-name",
				ReleaseVersion: "release-version",
			},
			desiredState: &ChartState{
				ReleaseName: "release-name",
				ChannelName: "channel-name",
				ChartValues: map[string]interface{}{
					"key": "value",
				},
				ReleaseVersion: "release-version",
			},
			expectedUpdateState: &ChartState{
				ReleaseName: "release-name",
				ChannelName: "channel-name",
				ChartValues: map[string]interface{}{
					"key": "value",
				},
				ReleaseVersion: "release-version",
			},
		},
		{
			description: "case 4: nonempty current state, desired state has different values, expected desired state",
			currentState: &ChartState{
				ReleaseName:    "release-name",
				ChannelName:    "channel-name",
				ReleaseVersion: "release-version",
			},
			desiredState: &ChartState{
				ReleaseName: "release-name",
				ChannelName: "channel-name",
				ChartValues: map[string]interface{}{
					"key": "new-value",
				},
				ReleaseVersion: "release-version",
			},
			expectedUpdateState: &ChartState{
				ReleaseName: "release-name",
				ChannelName: "channel-name",
				ChartValues: map[string]interface{}{
					"key": "new-value",
				},
				ReleaseVersion: "release-version",
			},
		},
		{
			description: "case 5: current state has values, desired state has equal values, empty update change",
			currentState: &ChartState{
				ReleaseName: "release-name",
				ChannelName: "channel-name",
				ChartValues: map[string]interface{}{
					"key": "value",
				},
				ReleaseVersion: "release-version",
			},
			desiredState: &ChartState{
				ReleaseName: "release-name",
				ChannelName: "channel-name",
				ChartValues: map[string]interface{}{
					"key": "value",
				},
				ReleaseVersion: "release-version",
			},
			expectedUpdateState: nil,
		},
		{
			description: "case 6: current state has values, desired state has new release and equal values, expected desired state",
			currentState: &ChartState{
				ReleaseName: "release-name",
				ChannelName: "channel-name",
				ChartValues: map[string]interface{}{
					"key": "value",
				},
				ReleaseVersion: "release-version",
			},
			desiredState: &ChartState{
				ReleaseName: "release-name",
				ChannelName: "channel-name",
				ChartValues: map[string]interface{}{
					"key": "value",
				},
				ReleaseVersion: "new-release-version",
			},
			expectedUpdateState: &ChartState{
				ReleaseName: "release-name",
				ChannelName: "channel-name",
				ChartValues: map[string]interface{}{
					"key": "value",
				},
				ReleaseVersion: "new-release-version",
			},
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

	for _, tc := range tcs {
		t.Run(tc.description, func(t *testing.T) {
			result, err := newResource.newUpdateChange(context.TODO(), customObject, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			if tc.expectedUpdateState == nil && result != nil {
				t.Fatal("expected", nil, "got", result)
			}
			if result != nil {
				updateChange, ok := result.(*ChartState)
				if !ok {
					t.Fatalf("expected '%T', got '%T'", updateChange, result)
				}
				if updateChange.ReleaseName != tc.expectedUpdateState.ReleaseName {
					t.Fatalf("expected ReleaseName %q, got %q", tc.expectedUpdateState.ReleaseName, updateChange.ReleaseName)
				}
				if updateChange.ChannelName != tc.expectedUpdateState.ChannelName {
					t.Fatalf("expected ChannelName %q, got %q", tc.expectedUpdateState.ChannelName, updateChange.ChannelName)
				}
			}
		})
	}
}
