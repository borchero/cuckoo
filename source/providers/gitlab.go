package providers

import (
	"errors"
	"fmt"

	"github.com/xanzy/go-gitlab"
)

// GitlabProject represents a single Gitlab project and provides information about it.
type GitlabProject struct {
	id     string
	client *gitlab.Client
}

// NewGitlabProject initializes a new GitLab project from the provided metadata.
func NewGitlabProject(serverHost, projectID, user, password string) (*GitlabProject, error) {
	// 1) Get client
	client, err := gitlab.NewBasicAuthClient(
		nil, fmt.Sprintf("https://%s", serverHost), user, password,
	)
	if err != nil {
		return nil, err
	}

	// 2) Get all repositories
	return &GitlabProject{
		id:     projectID,
		client: client,
	}, nil
}

// GetLatestTag returns the latest tag on the master branch.
func (project *GitlabProject) GetLatestTag() (string, error) {
	options := &gitlab.ListTagsOptions{}
	options.PerPage = 1
	tags, _, err := project.client.Tags.ListTags(project.id, options)
	if err != nil {
		return "", nil
	}

	if len(tags) == 0 {
		return "", errors.New("No tag found")
	}

	return tags[0].Name, nil
}
