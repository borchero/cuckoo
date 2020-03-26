package cdn

// Provider is an interface that enables accessing a content delivery network of a cloud provider.
type Provider interface {

	// Invalidate invalidates the given cache path and returns an error upon failure.
	Invalidate(path string) error
}
