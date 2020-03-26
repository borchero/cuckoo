package builder

import (
	"fmt"

	"go.borchero.com/cuckoo/utils"
)

type docker struct {
}

// NewDocker returns a new docker CLI instance to build images.
func NewDocker() Provider {
	return &docker{}
}

// Build performs the specified build using the Docker CLI.
func (docker *docker) Build(build Build) error {
	// 1) Build image with base image name
	args := []string{
		"build", "-t", build.Image,
		"-f", fmt.Sprintf("%s", build.Dockerfile),
	}
	for _, arg := range build.Args {
		args = append(args, "--build-arg", arg)
	}
	for _, secret := range build.Secrets {
		args = append(args, "--secret", secret)
	}
	if build.SSH {
		args = append(args, "--ssh", "default")
	}
	args = append(args, build.Context)

	err := utils.RunCommandWithEnv([]string{"DOCKER_BUILDKIT=1"}, "docker", args...)
	if err != nil {
		return err
	}

	// 2) Upload image to registry with all given tags (if tags are present)
	for _, tag := range build.Tags {
		fullImage := fmt.Sprintf("%s:%s", build.Image, tag)

		err := utils.RunCommand("docker", "tag", build.Image, fullImage)
		if err != nil {
			return err
		}

		err = utils.RunCommand("docker", "push", fullImage)
		if err != nil {
			return err
		}
	}

	return nil
}
