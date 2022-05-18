package release

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/helmclient/v4/pkg/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck

	"github.com/giantswarm/chart-operator/v2/service/internal/clientpair"
)

func Test_Resource_Release_newDeleteChange(t *testing.T) {
	testCases := []struct {
		obj                  v1alpha1.Chart
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
	{
		helmClients, err := clientpair.NewClientPair(clientpair.ClientPairConfig{
			Logger: microloggertest.New(),

			PrvHelmClient: helmclienttest.New(helmclienttest.Config{}),
			PubHelmClient: helmclienttest.New(helmclienttest.Config{}),
		})
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		c := Config{
			Fs:          afero.NewMemMapFs(),
			CtrlClient:  fake.NewFakeClient(), //nolint:staticcheck
			HelmClients: helmClients,
			K8sClient:   k8sfake.NewSimpleClientset(),
			Logger:      microloggertest.New(),

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
