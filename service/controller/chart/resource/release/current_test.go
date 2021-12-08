package release

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/helmclient/v4/pkg/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck

	"github.com/giantswarm/chart-operator/v2/service/controller/chart/controllercontext"
)

func Test_CurrentState(t *testing.T) {
	testCases := []struct {
		name           string
		obj            *v1alpha1.Chart
		releaseContent *helmclient.ReleaseContent
		returnedError  error
		expectedState  ReleaseState
		expectedError  bool
	}{
		{
			name: "case 0: basic match",
			obj: &v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/values-md5-checksum": "1ee001c5286ca00fdf64d9660c04bde2",
					},
				},
				Spec: v1alpha1.ChartSpec{
					Name: "prometheus",
				},
			},
			releaseContent: &helmclient.ReleaseContent{
				Name:   "prometheus",
				Status: "DEPLOYED",
				Values: map[string]interface{}{
					"key": "value",
				},
				Version: "0.1.2",
			},
			expectedState: ReleaseState{
				Name:              "prometheus",
				Status:            "DEPLOYED",
				ValuesMD5Checksum: "1ee001c5286ca00fdf64d9660c04bde2",
				Version:           "0.1.2",
			},
		},
		{
			name: "case 1: different values",
			obj: &v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/values-md5-checksum": "5eb63bbbe01eeed093cb22bb8f5acdc3",
					},
				},
				Spec: v1alpha1.ChartSpec{
					Name: "prometheus",
				},
			},
			releaseContent: &helmclient.ReleaseContent{
				Name:   "prometheus",
				Status: "FAILED",
				Values: map[string]interface{}{
					"key":     "value",
					"another": "value",
				},
				Version: "1.2.3",
			},
			expectedState: ReleaseState{
				Name:              "prometheus",
				Status:            "FAILED",
				ValuesMD5Checksum: "5eb63bbbe01eeed093cb22bb8f5acdc3",
				Version:           "1.2.3",
			},
		},
		{
			name: "case 2: empty state when error for no release present",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "prometheus",
				},
			},
			releaseContent: &helmclient.ReleaseContent{},
			returnedError:  fmt.Errorf("No such release: prometheus"),
			expectedError:  true,
		},
		{
			name: "case 3: unexpected error",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "prometheus",
				},
			},
			releaseContent: &helmclient.ReleaseContent{},
			returnedError:  fmt.Errorf("Unexpected error"),
			expectedError:  true,
		},
		{
			name: "case 4: chart cordoned",
			obj: &v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/cordon-reason": "testing upgrade",
						"chart-operator.giantswarm.io/cordon-until":  "2019-12-31T23:59:59Z",
					},
				},
				Spec: v1alpha1.ChartSpec{
					Name: "quay.io/giantswarm/chart-operator-chart",
				},
			},
			expectedState: ReleaseState{},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var ctx context.Context
			{
				c := controllercontext.Context{}
				ctx = controllercontext.NewContext(context.Background(), c)
			}

			var helmClient helmclient.Interface
			{
				c := helmclienttest.Config{
					DefaultReleaseContent: tc.releaseContent,
					DefaultError:          tc.returnedError,
				}
				helmClient = helmclienttest.New(c)
			}

			c := Config{
				Fs:         afero.NewMemMapFs(),
				CtrlClient: fake.NewFakeClient(),
				HelmClient: helmClient,
				K8sClient:  k8sfake.NewSimpleClientset(),
				Logger:     microloggertest.New(),

				TillerNamespace: "giantswarm",
			}

			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := r.GetCurrentState(ctx, tc.obj)
			switch {
			case err != nil && !tc.expectedError:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.expectedError:
				t.Fatalf("error == nil, want non-nil")
			}

			ReleaseState, err := toReleaseState(result)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			if !cmp.Equal(ReleaseState, tc.expectedState) {
				t.Fatalf("want matching ReleaseState \n %s", cmp.Diff(ReleaseState, tc.expectedState))
			}
		})
	}

}
