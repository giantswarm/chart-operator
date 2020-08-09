package release

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/apiextensions/v2/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apiextensions/v2/pkg/clientset/versioned/fake"
	"github.com/giantswarm/helmclient/v2/pkg/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func Test_Resource_Release_newDeleteChange(t *testing.T) {
	testCases := []struct {
		obj                  v1alpha1.ChartConfig
		currentState         *ReleaseState
		desiredState         *ReleaseState
		expectedDeleteChange *ReleaseState
		description          string
	}{
		{
			description:          "case 0: empty current and desired, expected empty",
			currentState:         &ReleaseState{},
			desiredState:         &ReleaseState{},
			expectedDeleteChange: nil,
		},
		{
			description: "case 1: non-empty current, empty desired, expected empty",
			currentState: &ReleaseState{
				Name: "current",
			},
			desiredState:         &ReleaseState{},
			expectedDeleteChange: nil,
		},

		{
			description:  "case 2: empty current, non-empty desired, expected empty",
			currentState: &ReleaseState{},
			desiredState: &ReleaseState{
				Name: "desired",
			},
			expectedDeleteChange: nil,
		},
		{
			description: "case 3: equal non-empty current and desired, expected desired",
			currentState: &ReleaseState{
				Name: "desired",
			},
			desiredState: &ReleaseState{
				Name: "desired",
			},
			expectedDeleteChange: &ReleaseState{
				Name: "desired",
			},
		},
		{
			description: "case 4: different non-empty current and desired, expected empty",
			currentState: &ReleaseState{
				Name: "current",
			},
			desiredState: &ReleaseState{
				Name: "desired",
			},
			expectedDeleteChange: nil,
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
			result, err := newResource.newDeleteChange(context.TODO(), tc.obj, tc.currentState, tc.desiredState)
			if err != nil {
				t.Fatal("expected", nil, "got", err)
			}
			if tc.expectedDeleteChange == nil && result != nil {
				t.Fatal("expected", nil, "got", result)
			}
			if result != nil {
				deleteChange, ok := result.(*ReleaseState)
				if !ok {
					t.Fatalf("expected '%T', got '%T'", deleteChange, result)
				}
				if deleteChange.Name != "" && deleteChange.Name != tc.expectedDeleteChange.Name {
					t.Fatalf("expected %s, got %s", tc.expectedDeleteChange, deleteChange.Name)
				}
			}
		})
	}

}
