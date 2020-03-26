package repository

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

// GitlabRepo describes a git repository on GitLab.
type GitlabRepo struct {
	id     string
	client *gitlab.Client
}

// NewGitlabRepo creates access to a new GitLab project with the specified ID.
func NewGitlabRepo(serverHost, user, password, projectID string) (*GitlabRepo, error) {
	client, err := gitlab.NewBasicAuthClient(
		nil, fmt.Sprintf("https://%s", serverHost), user, password,
	)
	if err != nil {
		return nil, err
	}

	return &GitlabRepo{projectID, client}, nil
}

// Branches returns a list of all of the repository's current branches.
func (repo *GitlabRepo) Branches() ([]string, error) {
	options := &gitlab.ListBranchesOptions{}
	options.PerPage = 100
	branches, _, err := repo.client.Branches.ListBranches(repo.id, options)
	if err != nil {
		return nil, err
	}

	result := make([]string, len(branches))
	for i, branch := range branches {
		result[i] = branch.Name
	}

	return result, nil
}
