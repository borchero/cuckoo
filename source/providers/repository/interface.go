package repository

// Provider describes a git repository backend, either local or remote (such as GitLab or
// Bitbucket).
type Provider interface {
	Branches() ([]string, error)
}
