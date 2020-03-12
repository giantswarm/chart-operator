package tillermigration

import (
	"context"
)

// EnsureDeleted ensures Tiller is installed and the latest version. This is so
// the release resource has a working tiller to connect to.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
