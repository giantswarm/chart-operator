// +build k8srequired

package basic

import (
	"testing"
)

const (
	chartOperatorValues = `cnr:
  address: http://localhost:5000`
)

func TestInstallOperator(t *testing.T) {
	err := f.InstallOperator("chart-operator", "chartconfig", chartOperatorValues, ":${CIRCLE_SHA1}")

	if err != nil {
		t.Fatalf("could not install operator %v", err)
	}
}
