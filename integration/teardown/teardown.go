// +build k8srequired

package teardown

import (
	"fmt"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
)

func Teardown(f *framework.Host) error {
	//items := []string{"cnr-server", "chart-operator", "chart-operator-resource"}
	items := []string{}
	for _, item := range items {
		cmd := fmt.Sprintf("delete %s --purge", item)
		err := framework.HelmCmd(cmd)
		if err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}
