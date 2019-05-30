package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func Test_HasForceUpgradeAnnotation(t *testing.T) {
	testCases := []struct {
		name           string
		input          v1alpha1.Chart
		expectedResult bool
	}{
		{
			name: "case 0: no annotations",
			input: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expectedResult: false,
		},
		{
			name: "case 1: other annotations",
			input: v1alpha1.Chart{
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
			input: v1alpha1.Chart{
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
			input: v1alpha1.Chart{
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
			input: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"chart-operator.giantswarm.io/force-helm-upgrade": "invalid",
					},
				},
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := HasForceUpgradeAnnotation(tc.input)

			if result != tc.expectedResult {
				t.Fatalf("HasForceUpgradeAnnotation == %t, want %t", result, tc.expectedResult)
			}
		})
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

func Test_ToCustomResource(t *testing.T) {
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
			result, err := ToCustomResource(tc.input)
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

func Test_ValuesMD5ChecksumAnnotation(t *testing.T) {
	expectedMD5Checksum := "1ee001c5286ca00fdf64d9660c04bde2"

	obj := v1alpha1.Chart{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				ValuesMD5ChecksumAnnotationName: "1ee001c5286ca00fdf64d9660c04bde2",
			},
		},
	}

	if ValuesMD5ChecksumAnnotation(obj) != expectedMD5Checksum {
		t.Fatalf("values md5 checksum %#q, want %#q", ValuesMD5ChecksumAnnotation(obj), expectedMD5Checksum)
	}
}

func Test_VersionLabel(t *testing.T) {
	testCases := []struct {
		name            string
		obj             v1alpha1.Chart
		expectedVersion string
	}{
		{
			name: "case 0: basic match",
			obj: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"chart-operator.giantswarm.io/version": "1.0.0",
					},
				},
			},
			expectedVersion: "1.0.0",
		},
		{
			name: "case 1: different version",
			obj: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"chart-operator.giantswarm.io/version": "2.0.0",
					},
				},
			},
			expectedVersion: "2.0.0",
		},
		{
			name: "case 2: incorrect label",
			obj: v1alpha1.Chart{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app-operator.giantswarm.io/version": "1.0.0",
					},
				},
			},
			expectedVersion: "",
		},
		{
			name:            "case 3: missing label",
			obj:             v1alpha1.Chart{},
			expectedVersion: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := VersionLabel(tc.obj)

			if result != tc.expectedVersion {
				t.Fatalf("Version label == %#q, want %#q", result, tc.expectedVersion)
			}
		})
	}
}
