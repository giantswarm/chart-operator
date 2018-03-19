// +build k8srequired

package integration

import (
	"log"
	"os"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/chart-operator/service/chartconfig/v1/appr"
)

var (
	f *framework.Framework
)

// TestMain allows us to have common setup and teardown steps that are run
// once for all the tests https://golang.org/pkg/testing/#hdr-Main.
func TestMain(m *testing.M) {
	var v int
	var err error
	f, err = framework.New()
	if err != nil {
		log.Printf("unexpected error: %v\n", err)
		os.Exit(1)
	}

	if err := f.SetUp(); err != nil {
		log.Printf("unexpected error: %v\n", err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
		f.TearDown()
	}

	os.Exit(v)
}

func TestApprGetReleaseVersion(t *testing.T) {
	l, err := micrologger.New(micrologger.Config{})
	if err != nil {
		t.Errorf("could not create logger %v", err)
	}

	c := appr.Config{
		Logger: l,

		Address:      "http://localhost:5000",
		Organization: "giantswarm",
	}

	a, err := appr.New(c)
	if err != nil {
		t.Errorf("could not create appr %v", err)
	}

	customObject := v1alpha1.ChartConfig{
		Spec: v1alpha1.ChartConfigSpec{
			Chart: v1alpha1.ChartConfigSpecChart{
				Name:    "test-chart",
				Channel: "3-2-beta",
			},
		},
	}

	expected := "3.2.1"
	actual, err := a.GetReleaseVersion(customObject)
	if err != nil {
		t.Errorf("could not get release %v", err)
	}

	if expected != actual {
		t.Errorf("release didn't match expected, want %q, got %q", expected, actual)
	}
}
