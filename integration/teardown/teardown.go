// +build k8srequired

package teardown

import (
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"k8s.io/helm/pkg/helm"
)

func Teardown(f *framework.Host, helmClient *helmclient.Client) error {
	items := []string{"cnr-server", "chart-operator", "chart-operator-resource"}

	for _, item := range items {
		err := helmClient.DeleteRelease(item, helm.DeletePurge(true))
		if err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}
