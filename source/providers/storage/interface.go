package storage

// TransferObject describes the local path and the storage path of an item.
type TransferObject struct {
	LocalPath  string
	BucketPath string
}

// Provider is an interface that enables accessing object storage buckets of different cloud
// providers.
type Provider interface {

	// Upload uploads all the given objects to the bucket (potentially creating new versions) and
	// returns an error if the upload of at least one object fails.
	Upload(objects ...TransferObject) error

	// Delete deletes the objects at the specified paths and returns an error if removal fails.
	Delete(objects ...string) error

	// List lists the objects in the bucket (recursively) and returns an error if listing fails for
	// some reason.
	List() ([]string, error)
}
