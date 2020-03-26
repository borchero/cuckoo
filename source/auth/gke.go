package auth

import (
	"context"
	"fmt"
	"os"

	container "cloud.google.com/go/container/apiv1"
	"github.com/kelseyhightower/envconfig"
	"go.borchero.com/cuckoo/utils"
	containerpb "google.golang.org/genproto/googleapis/container/v1"
)

type gkeAuth struct {
	Project string `envconfig:"PROJECT" required:"true"`
	Zone    string `envconfig:"ZONE" required:"true"`
	Cluster string `envconfig:"CLUSTER" required:"true"`
}

type kubeConfigValues struct {
	ClusterEndpoint string
	ClusterCA       string
}

// NewGKEAuth returns a component able to authenticate a Kubernetes client against a Kubernetes
// cluster from the Google Kubernetes Engine.
func NewGKEAuth() Component {
	return &gkeAuth{}
}

func (*gkeAuth) Name() string {
	return "Google Kubernetes Engine"
}

func (auth *gkeAuth) EnsureAccess() error {
	if err := envconfig.Process("GKE", auth); err != nil {
		return err
	}

	// 1) Get GKE client
	ctx := context.Background()
	client, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return fmt.Errorf("Cannot connect to GCP: %s", err)
	}

	// 2) Get cluster
	request := &containerpb.GetClusterRequest{Name: auth.clusterURI()}
	cluster, err := client.GetCluster(ctx, request)
	if err != nil {
		return fmt.Errorf("Cannot get cluster metadata: %s", err)
	}

	// 3) Write kube config
	err = os.MkdirAll(fmt.Sprintf("%s/.kube", home()), 0755)
	if err != nil {
		return fmt.Errorf("Cannot initialize kubeconfig directory: %s", err)
	}

	values := kubeConfigValues{
		ClusterEndpoint: fmt.Sprintf("https://%s", cluster.GetEndpoint()),
		ClusterCA:       cluster.GetMasterAuth().GetClusterCaCertificate(),
	}
	destination := fmt.Sprintf("%s/.kube/config", home())
	err = utils.PopulateTemplateWrite("gke-kubeconfig.yaml", destination, values)
	if err != nil {
		return fmt.Errorf("Cannot write kubeconfig: %s", err)
	}

	return nil
}

func (auth *gkeAuth) clusterURI() string {
	return fmt.Sprintf(
		"projects/%s/locations/%s/clusters/%s", auth.Project, auth.Zone, auth.Cluster,
	)
}
