package cmd

import (
	"github.com/spf13/cobra"
	"go.borchero.com/cuckoo/auth"
	"go.borchero.com/typewriter"
)

const authDescription = `
The auth command checks whether cuckoo is authenticated for multiple components. As CI pipelines
are most commonly authenticated using environment variables, all components follow that paradigm.
It is important that components do NOT GUARANTEE that access is granted: if keys are wrong or do
not grant sufficient scopes for subsequent operations, this command does NOT detect that. The
components that are currently checked are the following:

* SSH: Checks for the existence of SSH_PRIVATE_KEY and adds the key to the SSH daemon if found.
	Optionally, SSH_KNOWN_HOSTS may be given as a newline-separated list of hosts to add to the
	known hosts file.
* Docker: Checks for the existence of DOCKER_HOST, DOCKER_USER and DOCKER_PASSWORD and performs a
	login if all of them are found. To make it easier to use the GitLab registry, it is also valid
	to use CI_REGISTRY, CI_REGISTRY_USER and CI_REGISTRY_PASSWORD.
* Google Cloud Platform: Checks for the existence of GOOGLE_CREDENTIALS which must be a valid JSON
	key file for a service account. No actual login is performed.
* Amazon Web Services: Checks for the existence of AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY. As
	for the Google Cloud Platform, no actual login is performed.
* Google Kubernetes Engine: Checks for the existence of GKE_PROJECT, GKE_ZONE and GKE_CLUSTER and
	updates user's kubeconfig accordingly if all environment variables are found.

Note that when you are not running in a CI environment and none of these environment variables are
set, authentication will still work as you expect from any other command line tool. Hence, there is
no need to call this function when running cuckoo outside from a CI Docker container.
`

func init() {
	authCommand := &cobra.Command{
		Use:   "auth",
		Short: "Check authentication for all components to simplify debugging.",
		Long:  authDescription,
		Args:  cobra.ExactArgs(0),
		Run:   runAuth,
	}
	rootCmd.AddCommand(authCommand)
}

func runAuth(cmd *cobra.Command, args []string) {
	logger := typewriter.NewCLILogger()

	components := []auth.Component{
		auth.NewSSHAuth(),
		auth.NewDockerAuth(),
		auth.NewGCPAuth(),
		auth.NewGKEAuth(),
		auth.NewAWSAuth(),
	}

	for _, c := range components {
		if err := c.EnsureAccess(); err != nil {
			logger.Errorf("[ ] %s: %s", c.Name(), err)
		} else {
			logger.Infof("[x] %s", c.Name())
		}
	}

	logger.Success("Done 🎉")
}
