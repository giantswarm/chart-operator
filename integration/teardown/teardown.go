// +build k8srequired

package teardown

import (
	"fmt"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"k8s.io/helm/pkg/helm"
)

func Teardown(f *framework.Host, helmClient *helmclient.Client) error {
	items := []string{"cnr-server", "chart-operator", "apiextensions-chart-config-e2e"}

	for _, item := range items {
		err := helmClient.DeleteRelease(fmt.Sprintf("%s-%s", f.TargetNamespace(), item), helm.DeletePurge(true))
		if err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}
