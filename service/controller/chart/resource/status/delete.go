package status

import (
	"context"
)

// EnsureDeleted is not implemented for the status resource.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
