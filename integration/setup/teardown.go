package setup

import (
	"context"

	"github.com/giantswarm/microerror"
)

func teardown(ctx context.Context, config Config) error {
	// Clean operator components.
	err := config.HelmClient.DeleteRelease(ctx, "giantswarm-chart-operator")
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
