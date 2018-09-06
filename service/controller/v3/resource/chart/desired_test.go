package chart

import (
	"context"
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
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
		obj           *v1alpha1.ChartConfig
		configMap     *apiv1.ConfigMap
		userConfigMap *apiv1.ConfigMap
		secret        *apiv1.Secret
		expectedState ChartState
		errorMatcher  func(error) bool
	}{
		{
			name: "case 0: basic match",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "chart-operator-chart",
						Channel: "0-1-beta",
						Release: "chart-operator",
					},
				},
			},
			expectedState: ChartState{
				ChartName:      "chart-operator-chart",
				ChartValues:    map[string]interface{}{},
				ChannelName:    "0-1-beta",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.2",
			},
		},
		{
			name: "case 1: basic match with empty config map",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name: "chart-operator-chart",
						ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
						Channel: "0.1-beta",
						Release: "chart-operator",
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
			expectedState: ChartState{
				ChartName:      "chart-operator-chart",
				ChartValues:    map[string]interface{}{},
				ChannelName:    "0.1-beta",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.2",
			},
		},
		{
			name: "case 2: basic match with config map value",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name: "chart-operator-chart",
						ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
						Channel: "0-1-beta",
						Release: "chart-operator",
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
			expectedState: ChartState{
				ChartName: "chart-operator-chart",
				ChartValues: map[string]interface{}{
					"test": "test",
				},
				ChannelName:    "0-1-beta",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.2",
			},
		},
		{
			name: "case 3: config map with multiple values",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name: "chart-operator-chart",
						ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
						Channel: "0-1-beta",
						Release: "chart-operator",
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
			expectedState: ChartState{
				ChartName: "chart-operator-chart",
				ChartValues: map[string]interface{}{
					"provider": "azure",
					// Numeric values in JSON will be deserialized to a float64.
					"replicas": float64(2),
				},
				ChannelName:    "0-1-beta",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.2",
			},
		},
		{
			name: "case 4: config map not found",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name: "chart-operator-chart",
						ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
						Channel: "0-1-beta",
						Release: "chart-operator",
					},
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
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name: "chart-operator-chart",
						Secret: v1alpha1.ChartConfigSpecSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
						Channel: "0-1-beta",
						Release: "chart-operator",
					},
				},
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
			expectedState: ChartState{
				ChartName: "chart-operator-chart",
				ChartValues: map[string]interface{}{
					"test": "test",
				},
				ChannelName:    "0-1-beta",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.2",
			},
		},
		{
			name: "case 6: secret with multiple values",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name: "chart-operator-chart",
						Secret: v1alpha1.ChartConfigSpecSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
						Channel: "0-1-beta",
						Release: "chart-operator",
					},
				},
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
			expectedState: ChartState{
				ChartName: "chart-operator-chart",
				ChartValues: map[string]interface{}{
					"secretpassword": "admin",
					"secretnumber":   float64(2),
				},
				ChannelName:    "0-1-beta",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.2",
			},
		},
		{
			name: "case 7: secret not found",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name: "chart-operator-chart",
						Secret: v1alpha1.ChartConfigSpecSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
						Channel: "0-1-beta",
						Release: "chart-operator",
					},
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
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name: "chart-operator-chart",
						ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:      "chart-operator-values-configmap",
							Namespace: "giantswarm",
						},
						Secret: v1alpha1.ChartConfigSpecSecret{
							Name:      "chart-operator-values-secret",
							Namespace: "giantswarm",
						},
						Channel: "0-1-beta",
						Release: "chart-operator",
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
			errorMatcher: IsInvalidConfig,
		},
		{
			name: "case 9: user configmap overrides values",
			obj: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name: "chart-operator-chart",
						ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:      "values-configmap",
							Namespace: "kube-system",
						},
						UserConfigMap: v1alpha1.ChartConfigSpecConfigMap{
							Name:      "user-configmap",
							Namespace: "kube-system",
						},
						Channel: "0-1-beta",
						Release: "custom-chart",
					},
				},
			},
			configMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "values-configmap",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"values.json": `{ "values-key-1": "test-value", "values-key-2": 2 }`,
				},
			},
			userConfigMap: &apiv1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "user-configmap",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"user-key":   "test-value",
					"user-key-2": "test-value-2",
				},
			},
			expectedState: ChartState{
				ChartName: "chart-operator-chart",
				ChartValues: map[string]interface{}{
					"values-key-1": "test-value",
					"values-key-2": float64(2),
					"configmap": map[string]interface{}{
						"user-key":   "test-value",
						"user-key-2": "test-value-2",
					},
				},
				ChannelName:    "0-1-beta",
				ReleaseName:    "chart-operator",
				ReleaseVersion: "0.1.2",
			},
		},
	}

	apprClient := &apprMock{
		defaultReleaseVersion: "0.1.2",
	}
	helmClient := &helmMock{
		defaultError: nil,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objs := make([]runtime.Object, 0, 0)
			if tc.configMap != nil {
				objs = append(objs, tc.configMap)
			}
			if tc.userConfigMap != nil {
				objs = append(objs, tc.userConfigMap)
			}
			if tc.secret != nil {
				objs = append(objs, tc.secret)
			}

			c := Config{
				ApprClient: apprClient,
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

			chartState, err := toChartState(result)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			if len(chartState.ChartValues) != len(tc.expectedState.ChartValues) {
				t.Fatalf("ChartState.ChartValues == %d, want %d", len(chartState.ChartValues), len(tc.expectedState.ChartValues))
			}

			if !reflect.DeepEqual(chartState.ChartValues, tc.expectedState.ChartValues) {
				t.Fatalf("ChartState == %#v, want %#v", chartState.ChartValues, tc.expectedState.ChartValues)
			}
		})
	}

}

func Test_mergeValuesConfigMaps(t *testing.T) {
	testCases := []struct {
		name           string
		values         map[string]interface{}
		userValues     map[string]interface{}
		expectedValues map[string]interface{}
		errorMatcher   func(error) bool
	}{
		{
			name:           "case 0: both values empty",
			values:         map[string]interface{}{},
			userValues:     map[string]interface{}{},
			expectedValues: map[string]interface{}{},
		},
		{
			name: "case 1: values non-empty, user values empty",
			values: map[string]interface{}{
				"val-1": "val1",
				"val-2": "val2",
			},
			userValues: map[string]interface{}{},
			expectedValues: map[string]interface{}{
				"val-1": "val1",
				"val-2": "val2",
			},
		},
		{
			name: "case 2: values non-empty, user values non-empty",
			values: map[string]interface{}{
				"val-1": "val1",
				"val-2": "val2",
			},
			userValues: map[string]interface{}{
				"user-val-1": "userval1",
				"user-val-2": "userval2",
			},
			expectedValues: map[string]interface{}{
				"val-1": "val1",
				"val-2": "val2",
				"configmap": map[string]interface{}{
					"user-val-1": "userval1",
					"user-val-2": "userval2",
				},
			},
		},
		{
			name: "case 3: both non-empty but not intersecting",
			values: map[string]interface{}{
				"configmap": map[string]interface{}{
					"val-1": "val1",
					"val-2": "val2",
				},
			},
			userValues: map[string]interface{}{
				"user-val-1": "userval1",
				"user-val-2": "userval2",
			},
			expectedValues: map[string]interface{}{
				"configmap": map[string]interface{}{
					"val-1":      "val1",
					"val-2":      "val2",
					"user-val-1": "userval1",
					"user-val-2": "userval2",
				},
			},
		},
		{
			name: "case 4: user values override generated values",
			values: map[string]interface{}{
				"configmap": map[string]interface{}{
					"val-1": "val1",
					"val-2": "val2",
				},
			},
			userValues: map[string]interface{}{
				"val-1": "custom-val",
				"val-2": "custom-val",
			},
			expectedValues: map[string]interface{}{
				"configmap": map[string]interface{}{
					"val-1": "custom-val",
					"val-2": "custom-val",
				},
			},
		},
		{
			name: "case 5: both values with some user overrides",
			values: map[string]interface{}{
				"other-val-1": "val1",
				"other-val-2": "val2",
				"configmap": map[string]interface{}{
					"val-1": "val1",
					"val-2": "val2",
					"val-3": "val-3",
				},
			},
			userValues: map[string]interface{}{
				"val-1": "custom-val",
				"val-2": "custom-val",
			},
			expectedValues: map[string]interface{}{
				"other-val-1": "val1",
				"other-val-2": "val2",
				"configmap": map[string]interface{}{
					"val-1": "custom-val",
					"val-2": "custom-val",
					"val-3": "val-3",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mergeValuesConfigMaps(tc.values, tc.userValues)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(result, tc.expectedValues) {
				t.Fatalf("Map == %q, want %q", result, tc.expectedValues)
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
			errorMatcher: IsInvalidConfig,
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
