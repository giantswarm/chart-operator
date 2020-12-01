package status

import (
	"strconv"
	"testing"
	"time"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/to"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_StatusResource_equals(t *testing.T) {
	testCases := []struct {
		name    string
		statusA v1alpha1.ChartStatus
		statusB v1alpha1.ChartStatus
		equal   bool
	}{
		{
			name: "case 0: both equal",
			statusA: v1alpha1.ChartStatus{
				AppVersion: "1.0.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					LastDeployed: &metav1.Time{Time: time.Date(2020, 12, 1, 9, 0, 0, 0, time.UTC)},
					Revision:     to.IntP(1),
					Status:       "deployed",
				},
				Version: "2.1.0",
			},
			statusB: v1alpha1.ChartStatus{
				AppVersion: "1.0.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					LastDeployed: &metav1.Time{Time: time.Date(2020, 12, 1, 9, 0, 0, 0, time.UTC)},
					Revision:     to.IntP(1),
					Status:       "deployed",
				},
				Version: "2.1.0",
			},
			equal: true,
		},
		{
			name:    "case 1: both empty",
			statusA: v1alpha1.ChartStatus{},
			statusB: v1alpha1.ChartStatus{},
			equal:   true,
		},
		{
			name:    "case 2: empty A, non-empty B",
			statusA: v1alpha1.ChartStatus{},
			statusB: v1alpha1.ChartStatus{
				AppVersion: "1.0.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					Revision: to.IntP(1),
					Status:   "deployed",
				},
				Version: "2.1.0",
			},
			equal: false,
		},
		{
			name: "case 3: different version",
			statusA: v1alpha1.ChartStatus{
				AppVersion: "1.0.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					Revision: to.IntP(1),
					Status:   "deployed",
				},
				Version: "2.1.0",
			},
			statusB: v1alpha1.ChartStatus{
				AppVersion: "1.1.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					Revision: to.IntP(1),
					Status:   "deployed",
				},
				Version: "2.2.0",
			},
			equal: false,
		},
		{
			name: "case 4: different revision",
			statusA: v1alpha1.ChartStatus{
				AppVersion: "1.0.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					Revision: to.IntP(1),
					Status:   "deployed",
				},
				Version: "2.1.0",
			},
			statusB: v1alpha1.ChartStatus{
				AppVersion: "1.0.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					Revision: to.IntP(2),
					Status:   "deployed",
				},
				Version: "2.1.0",
			},
			equal: false,
		},
		{
			name: "case 5: different last deployd",
			statusA: v1alpha1.ChartStatus{
				AppVersion: "1.0.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					LastDeployed: &metav1.Time{Time: time.Date(2020, 12, 1, 9, 0, 0, 0, time.UTC)},
					Revision:     to.IntP(1),
					Status:       "deployed",
				},
				Version: "2.1.0",
			},
			statusB: v1alpha1.ChartStatus{
				AppVersion: "1.0.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					LastDeployed: &metav1.Time{Time: time.Date(2020, 12, 1, 12, 0, 0, 0, time.UTC)},
					Revision:     to.IntP(1),
					Status:       "deployed",
				},
				Version: "2.1.0",
			},
			equal: false,
		},
		{
			name: "case 6: last deployed different nanos same seconds",
			statusA: v1alpha1.ChartStatus{
				AppVersion: "1.0.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					LastDeployed: &metav1.Time{Time: time.Date(2020, 12, 1, 9, 0, 30, 0, time.UTC)},
					Revision:     to.IntP(1),
					Status:       "deployed",
				},
				Version: "2.1.0",
			},
			statusB: v1alpha1.ChartStatus{
				AppVersion: "1.0.0",
				Reason:     "",
				Release: v1alpha1.ChartStatusRelease{
					LastDeployed: &metav1.Time{Time: time.Date(2020, 12, 1, 9, 0, 30, 30, time.UTC)},
					Revision:     to.IntP(1),
					Status:       "deployed",
				},
				Version: "2.1.0",
			},
			equal: true,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			result := equals(tc.statusA, tc.statusB)
			if result != tc.equal {
				t.Fatalf("result == %t, want %t\n%s", result, tc.equal, cmp.Diff(tc.statusA, tc.statusB))
			}
		})
	}
}
