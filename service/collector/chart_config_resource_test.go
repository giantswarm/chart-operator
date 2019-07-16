package collector

import (
	"reflect"
	"testing"
	"time"
)

func Test_convertToTime(t *testing.T) {
	expectedTime, err := time.Parse(time.RFC3339, "2019-12-31T23:59:59Z")
	if err != nil {
		t.Errorf("time.Parse err = %v", err)
	}

	tests := []struct {
		name         string
		datetime     string
		expected     time.Time
		errorMatcher func(error) bool
	}{
		{
			name:     "case 1: normal timestamp parsing",
			datetime: "2019-12-31T23:59:59.000",
			expected: expectedTime,
		},
		{
			name:         "case 2: parsing error since unknown ",
			datetime:     "2019-12-31T23:59:59Z",
			errorMatcher: IsInvalidExecution,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := convertToTime(tc.datetime)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("convertToTime() = %v, want %v", got, tc.expected)
			}
		})
	}
}
