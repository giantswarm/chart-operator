package setup

import (
	"context"

	"github.com/giantswarm/chart-operator/integration/key"
	"github.com/giantswarm/microerror"
)

func teardown(ctx context.Context, config Config) error {
	// Clean operator components.
	err := config.Release.Delete(ctx, key.ChartOperatorReleaseName())
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
