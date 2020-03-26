package cmd

import (
	"github.com/spf13/cobra"
	"go.borchero.com/cuckoo/ci"
)

var env = ci.ReadEnvironment()

var rootCmd = &cobra.Command{
	Use:   "cuckoo",
	Short: "Efficient CI/CD for GitLab CI and Kubernetes.",
}

// Execute runs the root command of the CLI.
func Execute() error {
	return rootCmd.Execute()
}
