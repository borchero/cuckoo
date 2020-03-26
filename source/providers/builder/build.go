package builder

// Build identifies a particular build and wraps information about it.
type Build struct {
	Context    string
	Dockerfile string
	Image      string
	Tags       []string
	Args       map[string]string
	Secrets    []string
	SSH        bool
}
