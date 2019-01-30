package tiller

import (
	"context"
)

// EnsureDeleted is not implemented for the tiller resource.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
