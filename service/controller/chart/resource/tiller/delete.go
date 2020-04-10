package tiller

import (
	"context"

	"github.com/giantswarm/microerror"
)

// EnsureDeleted ensures Tiller is installed and the latest version. This is so
// the release resource has a working tiller to connect to.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	err := r.ensureTillerInstalled(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
