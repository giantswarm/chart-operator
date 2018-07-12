// +build k8srequired

package chartvalues

import (
	"fmt"
	"testing"

	"github.com/giantswarm/micrologger"
)

func TestChartValues(t *testing.T) {
	const cr = "chart-operator-resource"

	charts := []chart.Chart{
		{
			Channel: "1-0-beta",
			Release: "1.0.0",
			Tarball: "/e2e/fixtures/tb-chart-1.0.0.tgz",
			Name:    "tb-chart",
		},
	}

	// Setup
	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Fatalf("could not create logger %v", err)
	}

	gsHelmClient, err := createGsHelmClient()
	if err != nil {
		t.Fatalf("could not create giantswarm helmClient %v", err)
	}

	err = chart.Push(f, charts)
	if err != nil {
		t.Fatalf("could not push inital charts to cnr %v", err)
	}

	// Test Creation
	l.Log("level", "debug", "message", fmt.Sprintf("creating %s", cr))
	err = f.InstallResource(cr, templates.ChartOperatorResourceValues, ":stable")
	if err != nil {
		t.Fatalf("could not install %q %v", cr, err)
	}
}
