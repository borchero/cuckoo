package builder

// Provider is an interface adopted by a tool that is able to build Docker images in the OCI format.
type Provider interface {
	Build(build Build) error
}
