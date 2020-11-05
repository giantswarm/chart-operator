package release

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/clientset/versioned/fake"
	"github.com/giantswarm/helmclient/v3/pkg/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Release_newUpdateChange(t *testing.T) {
	testCases := []struct {
		name                string
		obj                 v1alpha1.Chart
		currentState        *ReleaseState
		desiredState        *ReleaseState
		expectedUpdateState *ReleaseState
	}{
		{
			name:         "case 0: empty current state, empty update change",
			currentState: &ReleaseState{},
			desiredState: &ReleaseState{
				Name: "desired-release-name",
			},
			expectedUpdateState: nil,
		},
		{
			name: "case 1: nonempty current state, equal desired state, empty update change",
			currentState: &ReleaseState{
				Name:    "release-name",
				Status:  "release-status",
				Version: "release-version",
			},
			desiredState: &ReleaseState{
				Name:    "release-name",
				Status:  "release-status",
				Version: "release-version",
			},
			expectedUpdateState: nil,
		},
		{
			name: "case 2: nonempty current state, different release version in desired state, expected desired state",
			currentState: &ReleaseState{
				Name:    "current-release-name",
				Version: "current-release-version",
			},
			desiredState: &ReleaseState{
				Name:    "desired-release-name",
				Version: "desired-release-version",
			},
			expectedUpdateState: &ReleaseState{
				Name: "desired-release-name",
			},
		},
		{
			name: "case 3: nonempty current state, desired state has values, expected desired state",
			currentState: &ReleaseState{
				Name:    "release-name",
				Version: "release-version",
			},
			desiredState: &ReleaseState{
				Name:              "release-name",
				ValuesMD5Checksum: "checksum",
				Version:           "release-version",
			},
			expectedUpdateState: &ReleaseState{
				Name:              "release-name",
				ValuesMD5Checksum: "checksum",
				Version:           "release-version",
			},
		},
		{
			name: "case 4: nonempty current state, desired state has different values, expected desired state",
			currentState: &ReleaseState{
				Name:              "release-name",
				ValuesMD5Checksum: "old-checksum",
				Version:           "release-version",
			},
			desiredState: &ReleaseState{
				Name:              "release-name",
				ValuesMD5Checksum: "new-checksum",
				Version:           "release-version",
			},
			expectedUpdateState: &ReleaseState{
				Name:              "release-name",
				ValuesMD5Checksum: "new-checksum",
				Version:           "release-version",
			},
		},
		{
			name: "case 5: current state has values, desired state has equal values, empty update change",
			currentState: &ReleaseState{
				Name:              "release-name",
				ValuesMD5Checksum: "checksum",
				Version:           "release-version",
			},
			desiredState: &ReleaseState{
				Name:              "release-name",
				ValuesMD5Checksum: "checksum",
				Version:           "release-version",
			},
			expectedUpdateState: nil,
		},
		{
			name: "case 6: current state has values, desired state has new release and equal values, expected desired state",
			currentState: &ReleaseState{
				Name:              "release-name",
				ValuesMD5Checksum: "checksum",
				Version:           "release-version",
			},
			desiredState: &ReleaseState{
				Name:              "release-name",
				ValuesMD5Checksum: "checksum",
				Version:           "new-release-version",
			},
			expectedUpdateState: &ReleaseState{
				Name:              "release-name",
				ValuesMD5Checksum: "checksum",
				Version:           "new-release-version",
			},
		},
		{
			name: "case 7: nonempty current state, desired state has different status, expected desired state",
			currentState: &ReleaseState{
				Name:    "release-name",
				Status:  "release-status",
				Version: "release-version",
			},
			desiredState: &ReleaseState{
				Name:    "release-name",
				Status:  "desired-status",
				Version: "release-version",
			},
			expectedUpdateState: &ReleaseState{
				Name:    "release-name",
				Status:  "desired-status",
				Version: "release-version",
			},
		},
	}
	var newResource *Resource
	var err error
	{
		c := Config{
			Fs:         afero.NewMemMapFs(),
			G8sClient:  fake.NewSimpleClientset(),
			HelmClient: helmclienttest.New(helmclienttest.Config{}),
			K8sClient:  k8sfake.NewSimpleClientset(),
			Logger:     microloggertest.New(),

			TillerNamespace: "giantswarm",
		}

		newResource, err = New(c)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, err := newResource.newUpdateChange(context.TODO(), &tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			if tc.expectedUpdateState == nil && result != nil {
				t.Fatal("expected", nil, "got", result)
			}
			if result != nil {
				updateChange, ok := result.(*ReleaseState)
				if !ok {
					t.Fatalf("expected '%T', got '%T'", updateChange, result)
				}
				if updateChange.Name != tc.expectedUpdateState.Name {
					t.Fatalf("expected Name %q, got %q", tc.expectedUpdateState.Name, updateChange.Name)
				}
				if updateChange.Status != tc.expectedUpdateState.Status {
					t.Fatalf("expected Status %q, got %q", tc.expectedUpdateState.Status, updateChange.Status)
				}
			}
		})
	}
}
