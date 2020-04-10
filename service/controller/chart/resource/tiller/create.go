package tiller

import (
	"context"

	"github.com/giantswarm/microerror"
)

// EnsureCreated ensures Tiller is installed and the latest version.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	err := r.ensureTillerInstalled(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
