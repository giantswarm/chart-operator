package release

import (
	"context"
	"strconv"
	"testing"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/helmclient/v4/pkg/helmclient"
	"github.com/giantswarm/helmclient/v4/pkg/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/chart-operator/v4/service/controller/chart/controllercontext"

	"github.com/giantswarm/chart-operator/v4/service/internal/clientpair"
)

func Test_DesiredState(t *testing.T) {
	testCases := []struct {
		name          string
		obj           *v1alpha1.Chart
		configMap     *apiv1.ConfigMap
		secret        *apiv1.Secret
		expectedState ReleaseState
		errorMatcher  func(error) bool
	}{
		{
			name: "case 0: basic match",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name:    "chart-operator-chart",
					Version: "0.1.2",
				},
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmclient.StatusDeployed,
				ValuesMD5Checksum: "",
				Values:            map[string]interface{}{},
				Version:           "0.1.2",
			},
		},
		{
			name: "case 1: basic match with empty config map",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
					},
					Version: "1.2.3",
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{},
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmclient.StatusDeployed,
				ValuesMD5Checksum: "",
				Values:            map[string]interface{}{},
				Version:           "1.2.3",
			},
		},
		{
			name: "case 2: basic match with config map value",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
					},
					Version: "0.1.2",
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": `test: test`,
				},
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmclient.StatusDeployed,
				ValuesMD5Checksum: "6e5ae9a10fd227006b0f938c51cb300b",
				Values: map[string]interface{}{
					"test": "test",
				},
				Version: "0.1.2",
			},
		},
		{
			name: "case 3: config map with multiple values",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
					},
					Version: "0.1.2",
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": `"provider": "azure"
"replicas": 2`},
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmclient.StatusDeployed,
				ValuesMD5Checksum: "4845bfb2cf922d7527886ac13599ea3b",
				Values: map[string]interface{}{
					"provider": "azure",
					"replicas": 2,
				},
				Version: "0.1.2",
			},
		},
		{
			name: "case 4: config map not found",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
					},
					Version: "0.1.2",
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-values-configmap",
					Namespace: "giantswarm",
				},
			},
			errorMatcher: IsNotFound,
		},
		{
			name: "case 5: basic match with secret value",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						Secret: v1alpha1.ChartSpecConfigSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
					},
					Version: "0.1.2",
				},
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-secret",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"values": []byte(`"test": "test"`),
				},
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmclient.StatusDeployed,
				ValuesMD5Checksum: "6e5ae9a10fd227006b0f938c51cb300b",
				Values: map[string]interface{}{
					"test": "test",
				},
				Version: "0.1.2",
			},
		},
		{
			name: "case 6: secret with multiple values",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						Secret: v1alpha1.ChartSpecConfigSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
					},
					Version: "0.1.2",
				},
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-secret",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"values": []byte(`"secretpassword": "admin"
"secretnumber": 2`),
				},
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmclient.StatusDeployed,
				ValuesMD5Checksum: "2187a8fce91c3765a74d462062af7526",
				Values: map[string]interface{}{
					"secretnumber":   2,
					"secretpassword": "admin",
				},
				Version: "0.1.2",
			},
		},
		{
			name: "case 7: secret not found",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						Secret: v1alpha1.ChartSpecConfigSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
					},
					Version: "0.1.2",
				},
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-values-secret",
					Namespace: "giantswarm",
				},
			},
			errorMatcher: IsNotFound,
		},
		{
			name: "case 8: secret and configmap clash",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
					Config: v1alpha1.ChartSpecConfig{
						ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
						Secret: v1alpha1.ChartSpecConfigSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
					},
					Version: "0.1.2",
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values": `"username": "admin"
"replicas": 2`,
				},
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-secret",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"values": []byte(`"username": "admin"
"secretnumber": 2
"floatnumber": 3.14`),
				},
			},
			expectedState: ReleaseState{
				Name:              "chart-operator-chart",
				Status:            helmclient.StatusDeployed,
				ValuesMD5Checksum: "3b8440387b1462ecdceb25c4cb9ff065",
				Values: map[string]interface{}{
					"replicas":     2,
					"secretnumber": 2,
					"floatnumber":  3.14,
					"username":     "admin",
				},
				Version: "0.1.2",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			objs := make([]runtime.Object, 0)
			if tc.configMap != nil {
				objs = append(objs, tc.configMap)
			}
			if tc.secret != nil {
				objs = append(objs, tc.secret)
			}

			var ctx context.Context
			{
				c := controllercontext.Context{}
				ctx = controllercontext.NewContext(context.Background(), c)
			}

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
				K8sClient:   k8sfake.NewClientset(objs...),
				Logger:      microloggertest.New(),

				TillerNamespace: "giantswarm",
			}
			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := r.GetDesiredState(ctx, tc.obj)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			releaseState, err := toReleaseState(result)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			desiredYAML, err := yaml.Marshal(releaseState.Values)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			expectedYAML, err := yaml.Marshal(tc.expectedState.Values)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			if !cmp.Equal(desiredYAML, expectedYAML) {
				t.Fatalf("want matching ValuesYAML \n %s", cmp.Diff(desiredYAML, expectedYAML))
			}

			if !cmp.Equal(releaseState, tc.expectedState) {
				t.Fatalf("want matching ReleaseState \n %s", cmp.Diff(releaseState, tc.expectedState))
			}
		})
	}
}
