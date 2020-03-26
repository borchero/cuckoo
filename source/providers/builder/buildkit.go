package builder

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/containerd/console"
	"github.com/moby/buildkit/client"
	_ "github.com/moby/buildkit/client/connhelper/dockercontainer" // connection docker-container
	_ "github.com/moby/buildkit/client/connhelper/kubepod"         // connection kube-pod
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/session/secrets/secretsprovider"
	"github.com/moby/buildkit/session/sshforward/sshprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	"go.borchero.com/typewriter"
	"golang.org/x/sync/errgroup"
)

type buildKit struct {
	buildkit *client.Client
	logger   typewriter.CLILogger
}

// NewBuildKit returns an instance which uses buildctl and connects to a remote buildkitd daemon
// to build images.
func NewBuildKit(host string, logger typewriter.CLILogger) (Provider, error) {
	ctx := context.Background()
	buildkit, err := client.New(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain Buildkit client: %s", err)
	}

	return &buildKit{buildkit, logger}, nil
}

// Build performs the specified build using BuildKit.
func (kit *buildKit) Build(build Build) error {
	// 1) Get Dockerfile folder
	dockerfile, cancel, err := kit.dockerfileFolder(build.Dockerfile)
	defer cancel()
	if err != nil {
		return err
	}

	// 2) Assemble request for BuildKit
	// 2.1) Authenticate for session
	// 2.1.1) Docker
	attachable := []session.Attachable{
		authprovider.NewDockerAuthProvider(os.Stderr),
	}

	// 2.1.2) Secrets
	if len(build.Secrets) > 0 {
		fileSources := make([]secretsprovider.FileSource, len(build.Secrets))
		for i, secret := range build.Secrets {
			for _, item := range strings.Split(secret, ",") {
				splits := strings.Split(item, "=")
				if len(splits) != 2 {
					return fmt.Errorf("Secret '%s' has a wrong format", item)
				}
				switch splits[0] {
				case "id":
					fileSources[i].ID = splits[1]
				case "src":
					fileSources[i].FilePath = splits[1]
				default:
					return fmt.Errorf("Unknown key '%s' for secret", splits[0])
				}
			}
		}

		store, err := secretsprovider.NewFileStore(fileSources)
		if err != nil {
			return fmt.Errorf("Failed to store secrets in file store: %s", err)
		}

		attachable = append(attachable, secretsprovider.NewSecretProvider(store))
	}

	// 2.1.3) SSH
	if build.SSH {
		configs := []sshprovider.AgentConfig{
			sshprovider.AgentConfig{ID: "default"},
		}

		provider, err := sshprovider.NewSSHAgentProvider(configs)
		if err != nil {
			return fmt.Errorf("Failed to get SSH agent provider: %s", err)
		}

		attachable = append(attachable, provider)
	}

	// 2.2) Build base solver
	solveOpt := client.SolveOpt{
		Frontend:      "dockerfile.v0",
		FrontendAttrs: map[string]string{},
		Session:       attachable,
		LocalDirs: map[string]string{
			"context":    build.Context,
			"dockerfile": dockerfile,
		},
	}

	// 2.3) Set build arguments
	for key, value := range build.Args {
		attribute := fmt.Sprintf("build-arg:%s", key)
		solveOpt.FrontendAttrs[attribute] = value
	}

	// 3) Build for each tag
	if len(build.Tags) == 0 {
		return kit.buildForTag("", solveOpt)
	}

	for _, tag := range build.Tags {
		img := fmt.Sprintf("%s:%s", build.Image, tag)
		if err := kit.buildForTag(img, solveOpt); err != nil {
			return err
		}
	}

	return nil
}

func (kit *buildKit) buildForTag(tag string, solveOpt client.SolveOpt) error {
	if tag != "" {
		solveOpt.Exports = []client.ExportEntry{
			client.ExportEntry{
				Type: "image",
				Attrs: map[string]string{
					"push": "true",
					"name": tag,
				},
			},
		}
	} else {
		solveOpt.Exports = []client.ExportEntry{}
	}

	ctx := context.Background()
	errGroup, ctx := errgroup.WithContext(ctx)
	ch := make(chan *client.SolveStatus)

	// 1) Solver
	errGroup.Go(func() error {
		if _, err := kit.buildkit.Solve(ctx, nil, solveOpt, ch); err != nil {
			return fmt.Errorf("Failed processing request: %s", err)
		}
		return nil
	})

	// 2) Progress
	errGroup.Go(func() error {
		console, err := console.ConsoleFromFile(os.Stderr)
		if err != nil {
			return fmt.Errorf("Failed getting console to print progress: %s", err)
		}
		return progressui.DisplaySolveStatus(context.Background(), "", console, os.Stderr, ch)
	})

	return errGroup.Wait()
}

// dockerfileFolder returns the folder for a Dockerfile and optionally a "cancel" function that
// destroys a temporary folder.
func (kit *buildKit) dockerfileFolder(dockerfile string) (string, func(), error) {
	dockerfileParts := strings.Split(dockerfile, "/")
	if dockerfileParts[len(dockerfileParts)-1] == "Dockerfile" {
		dir := strings.Join(dockerfileParts[:len(dockerfileParts)-1], "/")
		if dir == "" {
			dir = "."
		}
		return dir, func() {}, nil
	}

	dir, cleanup, err := kit.copyDockerfile(dockerfile)
	if err != nil {
		return "", cleanup, fmt.Errorf("Unable to copy Dockerfile to temporary directory: %s", err)
	}
	return dir, cleanup, nil
}

// copyDockerfile copies the Dockerfile at the given path to a temporary directory and renames it to
// "Dockerfile". It then returns the name of the temporary directory.
func (kit *buildKit) copyDockerfile(dockerfile string) (string, func(), error) {
	dir, err := ioutil.TempDir("", "cuckoo-*")
	if err != nil {
		return "", func() {}, fmt.Errorf("Failed to create temporary directory: %s", err)
	}

	cancel := func() {
		os.RemoveAll(dir)
	}

	source, err := os.Open(dockerfile)
	if err != nil {
		return "", cancel, fmt.Errorf("Failed opening existing Dockerfile: %s", err)
	}
	defer source.Close()

	target, err := os.Create(dockerfile + "/Dockerfile")
	if err != nil {
		return "", cancel, fmt.Errorf("Failed creating new Dockerfile: %s", err)
	}
	defer target.Close()

	if _, err := io.Copy(target, source); err != nil {
		return "", cancel, fmt.Errorf("Failed copying data: %s", err)
	}

	return dir, cancel, nil
}
