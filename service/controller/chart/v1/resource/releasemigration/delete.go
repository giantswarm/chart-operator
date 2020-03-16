package releasemigration

import (
	"context"
)

// EnsureDeleted is no-op method for this resource
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
