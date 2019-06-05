package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ChartName(t *testing.T) {
	expectedName := "chart-operator-chart"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
				Release: "chart-operator",
			},
		},
	}

	if ChartName(obj) != expectedName {
		t.Fatalf("chart name %s, want %s", ChartName(obj), expectedName)
	}
}

func Test_ChannelName(t *testing.T) {
	expectedChannel := "0.1-beta"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
				Release: "chart-operator",
			},
		},
	}

	if ChannelName(obj) != expectedChannel {
		t.Fatalf("channel name %s, want %s", ChannelName(obj), expectedChannel)
	}
}

func Test_ConfigMapName(t *testing.T) {
	expectedConfigMapName := "chart-operator-chart-values"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
				ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
					Name:      "chart-operator-chart-values",
					Namespace: "giantswarm",
				},
				Release: "chart-operator",
			},
		},
	}

	if ConfigMapName(obj) != expectedConfigMapName {
		t.Fatalf("config map name %s, want %s", ConfigMapName(obj), expectedConfigMapName)
	}
}

func Test_ConfigMapNamespace(t *testing.T) {
	expectedConfigMapNamespace := "giantswarm"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
				ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
					Name:      "chart-operator-chart-values",
					Namespace: "giantswarm",
				},
				Release: "chart-operator",
			},
		},
	}

	if ConfigMapNamespace(obj) != expectedConfigMapNamespace {
		t.Fatalf("config map namespace %s, want %s", ConfigMapNamespace(obj), expectedConfigMapNamespace)
	}
}

func Test_HasForceUpgradeAnnotation(t *testing.T) {
	testCases := []struct {
		name           string
		input          v1alpha1.ChartConfig
		expectedResult bool
		hasError       bool
	}{
		{
			name: "case 0: no annotations",
			input: v1alpha1.ChartConfig{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expectedResult: false,
		},
		{
			name: "case 1: other annotations",
			input: v1alpha1.ChartConfig{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"test": "test",
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "case 2: annotation present",
			input: v1alpha1.ChartConfig{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/force-helm-upgrade": "true",
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "case 3: annotation present but false",
			input: v1alpha1.ChartConfig{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/force-helm-upgrade": "false",
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "case 4: annotation present but invalid value",
			input: v1alpha1.ChartConfig{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/force-helm-upgrade": "invalid",
					},
				},
			},
			expectedResult: false,
			hasError:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := HasForceUpgradeAnnotation(tc.input)
			switch {
			case err != nil && tc.hasError == false:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.hasError == true:
				t.Fatalf("error == nil, want non-nil")
			}

			if result != tc.expectedResult {
				t.Fatalf("HasForceUpgradeAnnotation == %t, want %t", result, tc.expectedResult)
			}
		})
	}
}

func Test_Namespace(t *testing.T) {
	expected := "giantswarm"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Namespace: "giantswarm",
			},
		},
	}

	actual := Namespace(obj)
	if actual != expected {
		t.Fatalf("namespace %s, want %s", actual, expected)
	}
}

func Test_SecretName(t *testing.T) {
	expectedSecretName := "chart-operator-chart-secret"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
				Secret: v1alpha1.ChartConfigSpecSecret{
					Name:      "chart-operator-chart-secret",
					Namespace: "giantswarm",
				},
				Release: "chart-operator",
			},
		},
	}

	if SecretName(obj) != expectedSecretName {
		t.Fatalf("secret name %s, want %s", SecretName(obj), expectedSecretName)
	}
}

func Test_SecretNamespace(t *testing.T) {
	expectedSecretNamespace := "giantswarm"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
				Secret: v1alpha1.ChartConfigSpecSecret{
					Name:      "chart-operator-chart-secret",
					Namespace: "giantswarm",
				},
				Release: "chart-operator",
			},
		},
	}

	if SecretNamespace(obj) != expectedSecretNamespace {
		t.Fatalf("secret namespace %s, want %s", SecretNamespace(obj), expectedSecretNamespace)
	}
}

func Test_ReleaseName(t *testing.T) {
	expectedRelease := "chart-operator"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
				Release: "chart-operator",
			},
		},
	}

	if ReleaseName(obj) != expectedRelease {
		t.Fatalf("release name %s, want %s", ReleaseName(obj), expectedRelease)
	}
}

func Test_ReleaseStatus(t *testing.T) {
	expectedStatus := "DEPLOYED"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
				Release: "chart-operator",
			},
		},
		Status: v1alpha1.ChartConfigStatus{
			ReleaseStatus: "DEPLOYED",
		},
	}

	if ReleaseStatus(obj) != expectedStatus {
		t.Fatalf("release status %s, want %s", ReleaseStatus(obj), expectedStatus)
	}
}

func Test_ToCustomObject(t *testing.T) {
	testCases := []struct {
		name           string
		input          interface{}
		expectedObject v1alpha1.ChartConfig
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: basic match",
			input: &v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "chart-operator-chart",
						Channel: "0.1-beta",
						Release: "chart-operator",
					},
				},
			},
			expectedObject: v1alpha1.ChartConfig{
				Spec: v1alpha1.ChartConfigSpec{
					Chart: v1alpha1.ChartConfigSpecChart{
						Name:    "chart-operator-chart",
						Channel: "0.1-beta",
						Release: "chart-operator",
					},
				},
			},
		},
		{
			name:         "case 1: wrong type",
			input:        &v1alpha1.CertConfig{},
			errorMatcher: IsWrongTypeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ToCustomObject(tc.input)
			switch {
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case err != nil && !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !reflect.DeepEqual(result, tc.expectedObject) {
				t.Fatalf("Custom Object == %#v, want %#v", result, tc.expectedObject)
			}
		})
	}
}

func Test_UserConfigMapName(t *testing.T) {
	expectedConfigMapName := "chart-operator-user-values"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
				ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
					Name:      "chart-operator-chart-values",
					Namespace: "giantswarm",
				},
				Release: "chart-operator",
				UserConfigMap: v1alpha1.ChartConfigSpecConfigMap{
					Name:      "chart-operator-user-values",
					Namespace: "giantswarm",
				},
			},
		},
	}

	if UserConfigMapName(obj) != expectedConfigMapName {
		t.Fatalf("user config map name %s, want %s", UserConfigMapName(obj), expectedConfigMapName)
	}
}

func Test_UserConfigMapNamespace(t *testing.T) {
	expectedConfigMapNamespace := "default"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
				ConfigMap: v1alpha1.ChartConfigSpecConfigMap{
					Name:      "chart-operator-chart-values",
					Namespace: "giantswarm",
				},
				Release: "chart-operator",
				UserConfigMap: v1alpha1.ChartConfigSpecConfigMap{
					Name:      "chart-operator-custom-values",
					Namespace: "default",
				},
			},
		},
	}

	if UserConfigMapNamespace(obj) != expectedConfigMapNamespace {
		t.Fatalf("user config map namespace %s, want %s", UserConfigMapNamespace(obj), expectedConfigMapNamespace)
	}
}
