package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"go.borchero.com/cuckoo/utils"
	"go.borchero.com/typewriter"
)

const provisionDescription = `
The provision command performs a deployment of infrastructure using Hashicorp's Terraform.
Essentially, it runs the Terraform subcommands 'init', 'plan' and 'apply' all in one.
`

var provisionArgs struct {
	dryRun bool
}

func init() {
	provisionCommand := &cobra.Command{
		Use:   "provision",
		Short: "Provision infrastructure using Terraform.",
		Long:  provisionDescription,
		Args:  cobra.ExactArgs(0),
		Run:   runProvision,
	}

	provisionCommand.Flags().BoolVar(
		&provisionArgs.dryRun, "dry-run", false,
		"Only display changes without applying them.",
	)

	rootCmd.AddCommand(provisionCommand)
}

func runProvision(cmd *cobra.Command, args []string) {
	logger := typewriter.NewCLILogger()

	if !utils.ExecutableExists("terraform") {
		typewriter.Fail(logger, "Terraform executable does not exist", nil)
	}

	logger.Infof("Initializing temp dir...")
	err := os.MkdirAll("/tmp", 0755)
	if err != nil {
		typewriter.Fail(logger, "Failed to create tmp directory", err)
	}

	logger.Infof("Initializing backends...")
	err = utils.RunCommand("terraform", "init")
	if err != nil {
		typewriter.Fail(logger, "Failed to initialize", nil)
	}

	logger.Infof("Validating...")
	err = utils.RunCommand("terraform", "validate")
	if err != nil {
		typewriter.Fail(logger, "Invalid Terraform configuration", nil)
	}

	logger.Infof("Compute diff...")
	err = utils.RunCommand("terraform", "plan", "-out", ".terraform/planfile")
	if err != nil {
		typewriter.Fail(logger, "Failed to plan", nil)
	}

	if !provisionArgs.dryRun {
		logger.Infof("Apply changes...")
		err = utils.RunCommand("terraform", "apply", ".terraform/planfile")
		if err != nil {
			typewriter.Fail(logger, "Failed to apply changes", nil)
		}
	} else {
		logger.Infof("Skipping apply due to dry run")
	}

	logger.Success("Done 🎉")
}
