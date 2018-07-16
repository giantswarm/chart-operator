// +build k8srequired

package chartvalues

import (
	"fmt"
	"testing"

	"github.com/giantswarm/chart-operator/integration/chart"
	"github.com/giantswarm/chart-operator/integration/templates"
)

func TestChartValues(t *testing.T) {
	const cr = "apiextensions-chart-config-e2e"

	charts := []chart.Chart{
		{
			Channel: "1-0-beta",
			Release: "1.0.0",
			Tarball: "/e2e/fixtures/tb-chart-1.0.0.tgz",
			Name:    "tb-chart",
		},
	}

	err := chart.Push(f, charts)
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
