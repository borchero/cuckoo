package ci

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.borchero.com/cuckoo/providers"
)

var (
	semVer2Pattern        = regexp.MustCompile("^([0-9]+)\\.([0-9]+)\\.([0-9]+)$")
	relaxedSemVer2Pattern = regexp.MustCompile("^.*([0-9]+\\.[0-9]+\\.[0-9]+).*$")
)

// Manager exposes methods providing common functionality derived from CI variables.
type Manager struct {
	env           Environment
	gitlabProject *providers.GitlabProject
}

// NewManager creates a new manager without initializing any dependencies.
func NewManager(env Environment) *Manager {
	return &Manager{env: env}
}

// InitGitlabProject initiates a connection to a GitLab project by using the CI environment
// variables.
func (manager *Manager) InitGitlabProject() error {
	project, err := providers.NewGitlabProject(
		manager.env.GitlabHost, manager.env.Project.ID,
		manager.env.Registry.User, manager.env.Registry.Password,
	)
	if err != nil {
		return fmt.Errorf("Unable to initialize GitLab project: %s", err)
	}
	manager.gitlabProject = project
	return nil
}

// TagsFromTemplates takes in a list of templated tags and returns their expanded versions. If
// expanding fails, an error will be returned.
func (manager *Manager) TagsFromTemplates(templates []string) ([]string, error) {
	result := []string{}
	for _, template := range templates {
		if template == "%@" {
			for _, t := range []string{"%t", "%m", "%n", "latest"} {
				tag, err := manager.TagFromTemplate(t)
				if tag == "0" {
					continue
				}
				if err != nil {
					return nil, err
				}
				result = append(result, tag)
			}
		} else {
			tag, err := manager.TagFromTemplate(template)
			if tag == "0" {
				continue
			}
			if err != nil {
				return nil, err
			}
			result = append(result, tag)
		}
	}
	return result, nil
}

// TagFromTemplate expands a single templated tag and returns an error on failure.
func (manager *Manager) TagFromTemplate(template string) (string, error) {
	contains := strings.Contains

	// 1) Replace %t, %m and %n
	if contains(template, "%t") || contains(template, "%m") || contains(template, "%n") {
		var fullTag string
		if manager.env.Commit.Tag != "" {
			// 1.1) CI tag
			fullTag = manager.env.Commit.Tag
		} else {
			// 1.2) Ensure GitLab connection
			if manager.gitlabProject == nil {
				return "", errors.New(
					`Using templates %t, %m or %n within a tag requires the CI_COMMIT_TAG 
					environment variable or a connection to a GitLab repository. Specify 
					CI_SERVER_HOST, CI_PROJECT_ID, CI_REGISTRY_USER, and CI_REGISTRY_PASSWORD to 
					initiate such a connection`,
				)
			}

			// 1.3) Get latest tag
			var err error
			fullTag, err = manager.gitlabProject.GetLatestTag()
			if err != nil {
				return "", fmt.Errorf("Cannot fetch latest tag from Gitlab: %s", err)
			}
		}

		if !semVer2Pattern.MatchString(fullTag) {
			return "", fmt.Errorf("Tag '%s' does not follow SemVer2", fullTag)
		}

		// 1.4) Replace tag
		if contains(template, "%t") {
			template = strings.ReplaceAll(template, "%t", fullTag)
		}
		if contains(template, "%m") {
			template = strings.ReplaceAll(template, "%m", majorFromTag(fullTag))
		}
		if contains(template, "%n") {
			template = strings.ReplaceAll(template, "%n", majorMinorFromTag(fullTag))
		}
	}

	// 2) Replace %d
	if contains(template, "%d") {
		template = strings.ReplaceAll(template, "%d", time.Now().Format("2006-01-02"))
	}

	// 3) Replace %h
	if contains(template, "%h") {
		if manager.env.Commit.Hash == "" {
			return "", errors.New("Cannot use template %h as CI_COMMIT_SHA is not set")
		}
		template = strings.ReplaceAll(template, "%h", manager.env.Commit.Hash[:7])
	}

	// 4) Replace %r
	if contains(template, "%r") {
		if manager.env.Commit.Branch == "" {
			return "", errors.New("Cannot use template %r as CI_COMMIT_REF_NAME is not set")
		}
		if !relaxedSemVer2Pattern.MatchString(manager.env.Commit.Branch) {
			return "", fmt.Errorf(
				"Cannot use template %%r as branch '%s' does not contain valid SemVer2",
				manager.env.Commit.Branch,
			)
		}
		matches := relaxedSemVer2Pattern.FindStringSubmatch(manager.env.Commit.Branch)
		template = strings.ReplaceAll(template, "%r", matches[1])
	}

	return template, nil
}

// ImageNameFromTemplate returns the image path by replacing template values with values from the CI
// environment.
func (manager *Manager) ImageNameFromTemplate(template string) (string, error) {
	if strings.Contains(template, "%r") {
		if manager.env.Registry.Host == "" {
			return "", errors.New("Cannot use template %r as CI_REGISTRY is not set")
		}
		template = strings.ReplaceAll(template, "%r", manager.env.Registry.Host)
	}

	if strings.Contains(template, "%p") {
		if manager.env.Project.Path == "" {
			return "", errors.New("Cannot use template %p as CI_PROJECT_PATH is not set")
		}
		template = strings.ReplaceAll(template, "%p", manager.env.Project.Path)
	}

	return template, nil
}

func majorFromTag(tag string) string {
	matches := semVer2Pattern.FindStringSubmatch(tag)
	return matches[1]
}

func majorMinorFromTag(tag string) string {
	matches := semVer2Pattern.FindStringSubmatch(tag)
	return fmt.Sprintf("%s.%s", matches[1], matches[2])
}
