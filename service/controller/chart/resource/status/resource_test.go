package status

import (
	"strconv"
	"testing"
	"time"

	"github.com/giantswarm/to"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
)

func Test_Resource_Status_equals(t *testing.T) {
	testCases := []struct {
		name    string
		statusA v1alpha1.ChartStatus
		statusB v1alpha1.ChartStatus
		equals  bool
	}{
		{
			name:    "case 0: empty statuses match",
			statusA: v1alpha1.ChartStatus{},
			statusB: v1alpha1.ChartStatus{},
			equals:  true,
		},
		{
			name: "case 1: A non-empty and B empty",
			statusA: v1alpha1.ChartStatus{
				AppVersion: "v1.0.1",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					Revision: to.IntP(8),
					Status:   "deployed",
				},
				Version: "1.7.0",
			},
			statusB: v1alpha1.ChartStatus{},
			equals:  false,
		},
		{
			name: "case 2: A non-empty and B empty",
			statusA: v1alpha1.ChartStatus{
				AppVersion: "v1.0.1",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					LastDeployed: &metav1.Time{Time: time.Date(2021, time.January, 1, 21, 0, 0, 0, time.UTC)},
					Revision:     to.IntP(8),
					Status:       "deployed",
				},
				Version: "1.7.0",
			},
			statusB: v1alpha1.ChartStatus{},
			equals:  false,
		},
		{
			name: "case 3: A and B both non-empty",
			statusA: v1alpha1.ChartStatus{
				AppVersion: "v1.0.1",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					LastDeployed: &metav1.Time{Time: time.Date(2021, time.January, 1, 21, 0, 0, 0, time.UTC)},
					Revision:     to.IntP(8),
					Status:       "deployed",
				},
				Version: "1.7.0",
			},
			statusB: v1alpha1.ChartStatus{
				AppVersion: "v1.0.1",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					LastDeployed: &metav1.Time{Time: time.Date(2021, time.January, 1, 21, 0, 0, 0, time.UTC)},
					Revision:     to.IntP(8),
					Status:       "deployed",
				},
				Version: "1.7.0",
			},
			equals: false,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result := equals(tc.statusA, tc.statusB)

			if result != tc.equals {
				t.Fatalf("expected %t, got %t diff %s", tc.equals, result, cmp.Diff(tc.statusA, tc.statusB))
			}
		})
	}
}

/*
  	AppVersion: "v1.0.1",
  	Reason:     "",
  	Release: v1alpha1.ChartStatusRelease{
- 		LastDeployed: s"2020-11-26 15:00:46.59756787 +0000 UTC",
+ 		LastDeployed: s"2020-11-26 15:00:46 +0000 UTC",
  		Revision:     &8,
  		Status:       "deployed",
  	},
  	Version: "1.7.0",
*/
