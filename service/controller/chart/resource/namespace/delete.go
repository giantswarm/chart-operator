package namespace

import "context"

// EnsureDeleted is a no-op because the namespace for the app could be used for other apps, too.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
