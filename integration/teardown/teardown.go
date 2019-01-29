// +build k8srequired

package teardown

import (
	"context"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"k8s.io/helm/pkg/helm"

	"github.com/giantswarm/chart-operator/integration/env"
)

func Teardown(f *framework.Host, helmClient *helmclient.Client) error {
	// Clean operator components.
	err := framework.HelmCmd("delete --purge giantswarm-chart-operator")
	if err != nil {
		return microerror.Mask(err)
	}

	if env.TestedCustomResource() == env.ChartConfigCustomResource {
		// Clean chartconfig related components.
		items := []string{"cnr-server", "apiextensions-chart-config-e2e"}

		for _, item := range items {
			err := helmClient.DeleteRelease(context.TODO(), item, helm.DeletePurge(true))
			if err != nil {
				return microerror.Mask(err)
			}
		}
	}

	return nil
}
