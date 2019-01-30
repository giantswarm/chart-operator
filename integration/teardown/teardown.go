// +build k8srequired

package teardown

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
)

func Teardown(f *framework.Host, helmClient *helmclient.Client) error {
	// Clean operator components.
	err := framework.HelmCmd("delete --purge giantswarm-chart-operator")
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
