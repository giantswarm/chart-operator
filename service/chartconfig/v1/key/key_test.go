package key

import (
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
)

func Test_ChartName(t *testing.T) {
	expectedName := "chart-operator-chart"

	obj := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "chart-operator-chart",
				Channel: "0.1-beta",
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
			},
		},
	}

	if ChannelName(obj) != expectedChannel {
		t.Fatalf("chart name %s, want %s", ChannelName(obj), expectedChannel)
	}
}
