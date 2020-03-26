package ci

import (
	"github.com/kelseyhightower/envconfig"
	"go.borchero.com/cuckoo/utils"
)

// Environment provides all environment variables that can be deduced from the GitLab CI environment
// variables.
type Environment struct {
	Commit     EnvCommit
	Project    EnvProject
	Registry   EnvRegistry
	GitlabHost string `envconfig:"CI_SERVER_HOST"`
}

// EnvCommit wraps CI information about the most recent commit.
type EnvCommit struct {
	Branch string `envconfig:"CI_COMMIT_REF_NAME"`
	Tag    string `envconfig:"CI_COMMIT_TAG"`
	Hash   string `envconfig:"CI_COMMIT_SHA"`
	Slug   string `envconfig:"CI_COMMIT_REF_SLUG"`
}

// EnvProject wraps CI information about the GitLab project.
type EnvProject struct {
	ID        string `envconfig:"CI_PROJECT_ID"`
	Path      string `envconfig:"CI_PROJECT_PATH"`
	Directory string `envconfig:"CI_PROJECT_DIR"`
	Slug      string `envconfig:"CI_PROJECT_PATH_SLUG"`
}

// EnvRegistry wraps CI information about the GitLab registry.
type EnvRegistry struct {
	Host     string `envconfig:"CI_REGISTRY"`
	Image    string `envconfig:"CI_REGISTRY_IMAGE"`
	User     string `envconfig:"CI_REGISTRY_USER"`
	Password string `envconfig:"CI_REGISTRY_PASSWORD"`
}

// ReadEnvironment returns the environment from the current context and panics on error.
func ReadEnvironment() Environment {
	utils.MapEnvs(map[string]string{
		"DOCKER_HOST":     "CI_REGISTRY",
		"DOCKER_USER":     "CI_REGISTRY_USER",
		"DOCKER_PASSWORD": "CI_REGISTRY_PASSWORD",
	})

	var result Environment
	envconfig.Process("", &result)
	return result
}
