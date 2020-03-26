package cmd

import (
	"github.com/spf13/cobra"
	"go.borchero.com/cuckoo/ci"
	"go.borchero.com/cuckoo/providers"
	"go.borchero.com/typewriter"
)

const deployDescription = `
The deploy command deploys a Helm chart to a Kubernetes cluster. A Helm chart may be defined in
multiple ways:

* Remote Charts: In this case, the --repo argument and the --chart argument must be given.
* Local Charts: In this case, only the --chart argument must be given. Although the 'template'
	folder must exist, there is no need for a Chart.yaml file to exist. Dependencies should be put
	in a 'dependencies.yaml' file in this case.
* Local Directories: In this case, only the --chart argument must be given and set to an arbitrary
	directory. It serves as a "bundle" for multiple Kubernetes manifests which do not require
	value files to be defined. This way, there is no need for an extra 'template' folder.
* Local Files: In this case, the --chart argument must be set to a particular file. Deploying a
	single file as Helm chart serves as an alternative for 'kubectl apply' and provides additional
	features such as rollbacks.

For (actual) local Helm charts, tag and image may be set, otherwise they are ignored. They
automatically override the values 'image.name' and 'image.tag' in the values.yaml file. Tags and
images may be templated in the same way as in the build command. Consult its documentation to read
about these template values.

Make sure to be authenticated for Kubernetes or run 'cuckoo auth' prior to calling this command to
write the kubeconfig file.
`

var deployArgs struct {
	repo      string
	chart     string
	version   string
	name      string
	values    []string
	namespace string
	image     string
	tag       string
	dryRun    bool
}

func init() {
	deployCommand := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a Helm chart to a Kubernetes cluster.",
		Long:  deployDescription,
		Args:  cobra.ExactArgs(0),
		Run:   runDeploy,
	}

	deployCommand.Flags().StringVar(
		&deployArgs.repo, "repo", "",
		"The URL to a Helm repository when using a remote chart.",
	)
	deployCommand.Flags().StringVar(
		&deployArgs.chart, "chart", "./deploy/helm",
		"The chart to deploy.",
	)
	deployCommand.Flags().StringVar(
		&deployArgs.version, "version", "0.0.0",
		"The version of the chart to deploy. Only relevant for remote charts.",
	)
	deployCommand.Flags().StringVar(
		&deployArgs.name, "name", env.Project.Slug,
		"The name of the Helm release.",
	)
	deployCommand.Flags().StringArrayVarP(
		&deployArgs.values, "values", "f", []string{},
		"A path to one or multiple value files to set values from.",
	)
	deployCommand.Flags().StringVarP(
		&deployArgs.namespace, "namespace", "n", "default",
		"The namespace for deployed resources.",
	)
	deployCommand.Flags().StringVar(
		&deployArgs.image, "image", "",
		"The path for the image to deploy.",
	)
	deployCommand.Flags().StringVarP(
		&deployArgs.tag, "tag", "t", "",
		"The tag of the image to use for deployment. Defines appVersion of local charts.",
	)
	deployCommand.Flags().BoolVar(
		&deployArgs.dryRun, "dry-run", false,
		"Whether to perform a dry-run (useful for testing the chart).",
	)

	rootCmd.AddCommand(deployCommand)
}

func runDeploy(cmd *cobra.Command, args []string) {
	logger := typewriter.NewCLILogger()
	manager := ci.NewManager(env)

	// 1) Configure Helm release
	release, err := providers.NewHelmRelease(
		deployArgs.repo, deployArgs.chart, deployArgs.version,
		deployArgs.name, deployArgs.namespace, logger,
	)
	if err != nil {
		typewriter.Fail(logger, "Failed to prepare deployment", err)
	}

	// 2) Get image and tag for local charts
	image := ""
	tag := ""
	if release.IsLocalChart() {
		image, err = manager.ImageNameFromTemplate(deployArgs.image)
		if err != nil {
			typewriter.Fail(logger, "Cannot use the specified image", err)
		}

		tag, err = manager.TagFromTemplate(deployArgs.tag)
		if err != nil {
			typewriter.Fail(logger, "Cannot use the specified tag", err)
		}
	}

	// 3) Run upgrade
	err = release.Upgrade(deployArgs.values, image, tag, deployArgs.dryRun)
	if err != nil {
		typewriter.Fail(logger, "Failed to deploy", err)
	}

	logger.Success("Done 🎉")
}
