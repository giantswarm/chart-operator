// +build k8srequired

package setup

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/chart-operator/integration/teardown"
)

func WrapTestMain(f *framework.Host, m *testing.M) {
	var v int
	var err error

	err = f.CreateNamespace("giantswarm")
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	err = resources(f)
	if err != nil {
		log.Printf("%#v\n", err)
		v = 1
	}

	if v == 0 {
		v = m.Run()
	}

	if os.Getenv("KEEP_RESOURCES") != "true" {
		// only do full teardown when not on CI
		if os.Getenv("CIRCLECI") != "true" {
			err := teardown.Teardown(f)
			if err != nil {
				log.Printf("%#v\n", err)
				v = 1
			}
			// TODO there should be error handling for the framework teardown.
			f.Teardown()
		}
	}

	os.Exit(v)
}

func resources(f *framework.Host) error {
	const chartOperatorValues = `cnr:
  address: http://localhost:5000
`

	err := f.InstallOperator("chart-operator", "chartconfig", chartOperatorValues, ":${CIRCLE_SHA1}")

	if err != nil {
		return microerror.Mask(err)
	}

	err = installChartOperatorResource(f)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func installChartOperatorResource(f *framework.Host) error {
	const chartOperatorResourceValues = `chart:
  name: "test-chart"
  channel: "3-2-beta"
  namespace: "default"
  release: "3.2.0"
`

	chartOperatorResourceValuesEnv := os.ExpandEnv(chartOperatorResourceValues)
	d := []byte(chartOperatorResourceValuesEnv)

	tmpfile, err := ioutil.TempFile("", "chart-operator-resource-values")
	if err != nil {
		return microerror.Mask(err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(d)
	if err != nil {
		return microerror.Mask(err)
	}
	err = tmpfile.Close()
	if err != nil {
		return microerror.Mask(err)
	}

	err = framework.HelmCmd("registry install quay.io/giantswarm/chart-operator-resource-chart:stable -- -n chart-operator-resource --values " + tmpfile.Name())
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
