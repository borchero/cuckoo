package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.borchero.com/cuckoo/ci"
	"go.borchero.com/cuckoo/providers/builder"
	"go.borchero.com/cuckoo/utils"
	"go.borchero.com/typewriter"
)

const buildDescription = `
The build command builds a Docker image either with the Docker CLI (using DOCKER_BUILDKIT=1) or
with BuildKit directly if a BuildKit host is defined.

After building, the image can optionally be uploaded to a repository with one or multiple tags. For
defining the destination path (i.e. registry & tags), multiple template parameters may be used. Some
template parameters are only available under certain conditions. If they are not available, building
will fail:

* %r: The Docker hsot, defined by DOCKER_HOST or CI_REGISTRY.
* %p: The base path from a GitLab repository, given by CI_PROJECT_PATH.
* %t: Either the tag found in CI_COMMIT_TAG or the most recent tag on this repository's master
	branch. At least one of them must be found, CI_COMMIT_TAG takes precedence. Must be a valid
	SemVer 2.0 tag.
* %m: Derived from %t, written as <major>. Fails when %t would fail. Ignored when <major> is 0.
* %n: Derived from %t, written as <major>.<minor>. Fails when %t would fail.
* %r: A valid SemVer2 tag, extracted from the current branch name. CI_COMMIT_REF_NAME must be
	available.
* %h: The hash of the current commit. CI_COMMIT_SHA must be available.
* %d: The current date, written as YYYY-MM-dd.
* %@: Will set all of the following tags: %t, %m, %n, latest.
`

var buildArgs struct {
	buildKitHost string
	context      string
	dockerfile   string
	args         []string
	image        string
	tags         []string
	secrets      []string
	ssh          bool
}

func init() {
	buildCommand := &cobra.Command{
		Use:   "build",
		Short: "Build a Docker image and optionally push with various tags.",
		Long:  buildDescription,
		Args:  cobra.ExactArgs(0),
		Run:   runBuild,
	}

	buildCommand.Flags().StringVar(
		&buildArgs.buildKitHost, "buildkit-host", os.Getenv("BUILDKIT_HOST"),
		"The host of the BuildKit daemon to use for building images.",
	)
	buildCommand.Flags().StringVar(
		&buildArgs.context, "context", ".",
		"The context to use for building the image.",
	)
	buildCommand.Flags().StringVarP(
		&buildArgs.dockerfile, "dockerfile", "f", "Dockerfile",
		"The dockerfile used for building.",
	)
	buildCommand.Flags().StringArrayVar(
		&buildArgs.args, "arg", []string{},
		"Arguments to set for the image build.",
	)
	buildCommand.Flags().StringVar(
		&buildArgs.image, "image", "unnamed",
		"The path of the image to push to. Ignored if no tag is set.",
	)
	buildCommand.Flags().StringArrayVarP(
		&buildArgs.tags, "tag", "t", []string{},
		"The templates for tags to apply to the image. No push is executed without tag.",
	)
	buildCommand.Flags().StringArrayVar(
		&buildArgs.secrets, "secret", []string{},
		"Secret file to expose to the build (id=<name>,src=<file>).",
	)
	buildCommand.Flags().BoolVar(
		&buildArgs.ssh, "ssh", false,
		"Whether to use the host's default SSH daemon to supply SSH keys.",
	)

	rootCmd.AddCommand(buildCommand)
}

func runBuild(cmd *cobra.Command, args []string) {
	logger := typewriter.NewCLILogger()
	manager := ci.NewManager(env)

	// 1) Choose build tool
	var buildTool builder.Provider
	if buildArgs.buildKitHost == "" && utils.ExecutableExists("docker") {
		// 1.1) Docker
		buildTool = builder.NewDocker()
	} else {
		if buildArgs.buildKitHost == "" {
			typewriter.Fail(
				logger, "Host for BuildKit is not set and Docker executable cannot be found", nil,
			)
		}

		// 1.2) BuildKit
		var err error
		buildTool, err = builder.NewBuildKit(buildArgs.buildKitHost, logger)
		if err != nil {
			typewriter.Fail(logger, "BuildKit cannot be initialized", err)
		}
	}

	// 2) Get builder
	// 2.1) Get components
	image, err := manager.ImageNameFromTemplate(buildArgs.image)
	if err != nil {
		typewriter.Fail(logger, "Cannot use the specified image", err)
	}

	tags, err := manager.TagsFromTemplates(buildArgs.tags)
	if err != nil {
		typewriter.Fail(logger, "Cannot use the specified set of tags", err)
	}

	bargs := make(map[string]string)
	for _, arg := range buildArgs.args {
		split := strings.Split(arg, "=")
		bargs[split[0]] = split[1]
	}

	// 2.2) Get build info
	buildInfo := builder.Build{
		Context:    buildArgs.context,
		Dockerfile: buildArgs.dockerfile,
		Image:      image,
		Tags:       tags,
		Args:       bargs,
		Secrets:    buildArgs.secrets,
		SSH:        buildArgs.ssh,
	}

	// 3) Finally build
	// 3.1) Print description of what to do
	logger.Info("About to perform build...")
	logger.Infof(" - context: %s", buildInfo.Context)
	logger.Infof(" - dockerfile: %s", buildInfo.Dockerfile)
	logger.Infof(" - image: %s", buildInfo.Image)
	logger.Infof(" - tags: [%s]", strings.Join(buildInfo.Tags, ", "))
	logger.Infof(" - args: [%s]", strings.Join(buildArgs.args, ", "))
	logger.Infof(" - ssh: %t", buildInfo.SSH)

	// 3.2) If we use SSH, ensure that $SSH_AUTH_SOCK is set
	if buildInfo.SSH && os.Getenv("SSH_AUTH_SOCK") == "" {
		if err := utils.RunCommand("eval", "`ssh-agent -s`"); err != nil {
			typewriter.Fail(logger, "Failed starting SSH agent", nil)
		}
	}

	// 3.3) Build
	if err := buildTool.Build(buildInfo); err != nil {
		typewriter.Fail(logger, "Build failed", err)
	}

	logger.Success("Done 🎉")
}
