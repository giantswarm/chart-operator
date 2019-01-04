package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
)

func Test_ConfigMapName(t *testing.T) {
	expectedConfigMapName := "prometheus-values"

	obj := v1alpha1.Chart{
		Spec: v1alpha1.ChartSpec{
			Config: v1alpha1.ChartSpecConfig{
				ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
					Name:      "prometheus-values",
					Namespace: "monitoring",
				},
			},
		},
	}

	if ConfigMapName(obj) != expectedConfigMapName {
		t.Fatalf("config map name %#q, want %#q", ConfigMapName(obj), expectedConfigMapName)
	}
}

func Test_ConfigMapNamespace(t *testing.T) {
	expectedConfigMapNamespace := "monitoring"

	obj := v1alpha1.Chart{
		Spec: v1alpha1.ChartSpec{
			Config: v1alpha1.ChartSpecConfig{
				ConfigMap: v1alpha1.ChartSpecConfigConfigMap{
					Name:      "prometheus-values",
					Namespace: "monitoring",
				},
			},
		},
	}

	if ConfigMapNamespace(obj) != expectedConfigMapNamespace {
		t.Fatalf("config map namespace %#q, want %#q", ConfigMapNamespace(obj), expectedConfigMapNamespace)
	}
}
func Test_ReleaseName(t *testing.T) {
	expectedRelease := "my-prometheus"

	obj := v1alpha1.Chart{
		Spec: v1alpha1.ChartSpec{
			Name: "my-prometheus",
		},
	}

	if ReleaseName(obj) != expectedRelease {
		t.Fatalf("release name %s, want %s", ReleaseName(obj), expectedRelease)
	}
}

func Test_SecretName(t *testing.T) {
	expectedSecretName := "prometheus-secret-values"

	obj := v1alpha1.Chart{
		Spec: v1alpha1.ChartSpec{
			Config: v1alpha1.ChartSpecConfig{
				Secret: v1alpha1.ChartSpecConfigSecret{
					Name:      "prometheus-secret-values",
					Namespace: "monitoring",
				},
			},
		},
	}

	if SecretName(obj) != expectedSecretName {
		t.Fatalf("secret name %#q, want %#q", SecretName(obj), expectedSecretName)
	}
}

func Test_SecretNamespace(t *testing.T) {
	expectedSecretNamespace := "monitoring"

	obj := v1alpha1.Chart{
		Spec: v1alpha1.ChartSpec{
			Config: v1alpha1.ChartSpecConfig{
				Secret: v1alpha1.ChartSpecConfigSecret{
					Name:      "prometheus-values",
					Namespace: "monitoring",
				},
			},
		},
	}

	if SecretNamespace(obj) != expectedSecretNamespace {
		t.Fatalf("secret namespace %#q, want %#q", SecretNamespace(obj), expectedSecretNamespace)
	}
}

func Test_TarballURL(t *testing.T) {
	expectedTarballURL := "https://path.to/chart-1.0.0.tgz"

	obj := v1alpha1.Chart{
		Spec: v1alpha1.ChartSpec{
			TarballURL: "https://path.to/chart-1.0.0.tgz",
		},
	}

	if TarballURL(obj) != expectedTarballURL {
		t.Fatalf("tarball url %#q, want %#q", SecretNamespace(obj), expectedTarballURL)
	}
}

func Test_ToCustomObject(t *testing.T) {
	testCases := []struct {
		name           string
		input          interface{}
		expectedObject v1alpha1.Chart
		errorMatcher   func(error) bool
	}{
		{
			name: "case 0: basic match",
			input: &v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name:       "prometheus-chart",
					Namespace:  "monitoring",
					TarballURL: "https://path.to/chart.tgz",
				},
			},
			expectedObject: v1alpha1.Chart{
				Spec: v1alpha1.ChartSpec{
					Name:       "prometheus-chart",
					Namespace:  "monitoring",
					TarballURL: "https://path.to/chart.tgz",
				},
			},
		},
		{
			name:         "case 1: wrong type",
			input:        &v1alpha1.App{},
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
