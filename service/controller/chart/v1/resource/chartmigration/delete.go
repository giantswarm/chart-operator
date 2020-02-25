package chartmigration

import (
	"context"
)

// EnsureDeleted is not implemented as no migration is required.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
