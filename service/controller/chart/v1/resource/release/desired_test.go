package release

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/helmclient"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/helmclient/helmclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_DesiredState(t *testing.T) {
	testCases := []struct {
		name          string
		obj           *v1alpha1.Chart
		configMap     *apiv1.ConfigMap
		helmChart     helmclient.Chart
		secret        *apiv1.Secret
		expectedState ReleaseState
		errorMatcher  func(error) bool
	}{
		{
			name: "case 0: basic match",
			obj: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name: "chart-operator-chart",
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			expectedState: ReleaseState{
				Name:    "chart-operator-chart",
				Status:  helmDeployedStatus,
				Values:  map[string]interface{}{},
				Version: "0.1.2",
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
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{},
			},
			helmChart: helmclient.Chart{
				Version: "1.2.3",
			},
			expectedState: ReleaseState{
				Name:    "chart-operator-chart",
				Status:  helmDeployedStatus,
				Values:  map[string]interface{}{},
				Version: "1.2.3",
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
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values.json": `{ "test": "test" }`,
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			expectedState: ReleaseState{
				Name:   "chart-operator-chart",
				Status: helmDeployedStatus,
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
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values.json": `{ "provider": "azure", "replicas": 2 }`,
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			expectedState: ReleaseState{
				Name:   "chart-operator-chart",
				Status: helmDeployedStatus,
				Values: map[string]interface{}{
					"provider": "azure",
					// Numeric values in JSON will be deserialized to a float64.
					"replicas": float64(2),
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
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "missing-values-configmap",
					Namespace: "giantswarm",
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
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
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-secret",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"secret.json": []byte(`{ "test": "test" }`),
				},
			},
			expectedState: ReleaseState{
				Name:   "chart-operator-chart",
				Status: helmDeployedStatus,
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
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-secret",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"secret.json": []byte(`{ "secretpassword": "admin", "secretnumber": 2 }`),
				},
			},
			expectedState: ReleaseState{
				Name:   "chart-operator-chart",
				Status: helmDeployedStatus,
				Values: map[string]interface{}{
					"secretpassword": "admin",
					"secretnumber":   float64(2),
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
				},
			},
			helmChart: helmclient.Chart{
				Version: "0.1.2",
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
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-configmap",
					Namespace: "giantswarm",
				},
				Data: map[string]string{
					"values.json": `{ "username": "admin", "replicas": 2 }`,
				},
			},
			secret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "chart-operator-values-secret",
					Namespace: "giantswarm",
				},
				Data: map[string][]byte{
					"secret.json": []byte(`{ "username": "admin", "secretnumber": 2 }`),
				},
			},
			errorMatcher: IsInvalidExecution,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0, 0)
			if tc.configMap != nil {
				objs = append(objs, tc.configMap)
			}
			if tc.secret != nil {
				objs = append(objs, tc.secret)
			}

			var helmClient helmclient.Interface
			{
				c := helmclienttest.Config{
					LoadChartResponse: tc.helmChart,
				}
				helmClient = helmclienttest.New(c)
			}

			c := Config{
				Fs:         afero.NewMemMapFs(),
				HelmClient: helmClient,
				K8sClient:  fake.NewSimpleClientset(objs...),
				Logger:     microloggertest.New(),
			}
			r, err := New(c)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			result, err := r.GetDesiredState(context.TODO(), tc.obj)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			ReleaseState, err := toReleaseState(result)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			if !reflect.DeepEqual(ReleaseState, tc.expectedState) {
				t.Fatalf("ReleaseState == %#v, want %#v", ReleaseState, tc.expectedState)
			}
		})
	}

}

func Test_union(t *testing.T) {
	testCases := []struct {
		name         string
		inputA       map[string]interface{}
		inputB       map[string]interface{}
		expectedMap  map[string]interface{}
		errorMatcher func(error) bool
	}{
		{
			name: "case 0: both maps with exclusive entries",
			inputA: map[string]interface{}{
				"secret": "secret",
			},
			inputB: map[string]interface{}{
				"config": "config",
			},
			expectedMap: map[string]interface{}{
				"secret": "secret",
				"config": "config",
			},
		},
		{
			name: "case 1: only the first input",
			inputA: map[string]interface{}{
				"secret": "secret",
			},
			inputB: nil,
			expectedMap: map[string]interface{}{
				"secret": "secret",
			},
		},
		{
			name:   "case 2: only the second input",
			inputA: nil,
			inputB: map[string]interface{}{
				"config": "config",
			},
			expectedMap: map[string]interface{}{
				"config": "config",
			},
		},
		{
			name:        "case 3: no input",
			inputA:      nil,
			inputB:      nil,
			expectedMap: nil,
		},
		{
			name: "case 4: entries are not exclusive",
			inputA: map[string]interface{}{
				"secret": "secret",
			},
			inputB: map[string]interface{}{
				"config": "config",
				"secret": "config",
			},
			errorMatcher: IsInvalidExecution,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := union(tc.inputA, tc.inputB)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(result, tc.expectedMap) {
				t.Fatalf("Map == %q, want %q", result, tc.expectedMap)
			}
		})
	}
}
