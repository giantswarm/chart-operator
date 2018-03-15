package chart

import (
	"testing"
)

func Test_toChartState(t *testing.T) {
	testCases := []struct {
		name          string
		input         interface{}
		expectedState ChartState
		errorMatcher  func(error) bool
	}{
		{
			name: "case 0: basic match",
			input: &ChartState{
				ChartName:      "test-chart",
				ChannelName:    "test-channel",
				ReleaseVersion: "test-release",
			},
			expectedState: ChartState{
				ChartName:      "test-chart",
				ChannelName:    "test-channel",
				ReleaseVersion: "test-release",
			},
		},
		{
			name: "case 1: wrong type",
			input: ChartState{
				ChartName:      "test-chart",
				ChannelName:    "test-channel",
				ReleaseVersion: "test-release",
			},
			errorMatcher: IsWrongType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toChartState(tc.input)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if result != tc.expectedState {
				t.Fatalf("ChartState == %q, want %q", result, tc.expectedState)
			}
		})
	}
}
