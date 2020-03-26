package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kelseyhightower/envconfig"
	"go.borchero.com/cuckoo/utils"
)

type dockerAuth struct {
	Host     string `envconfig:"HOST" required:"true"`
	User     string `envconfig:"USER" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
}

type dockerAuthFile struct {
	Auths map[string]dockerAuthItem `json:"auths"`
}

type dockerAuthItem struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
}

// NewDockerAuth returns a component able to authenticate against a Docker registry.
func NewDockerAuth() Component {
	return &dockerAuth{}
}

func (*dockerAuth) Name() string {
	return "Docker Registry"
}

func (auth *dockerAuth) EnsureAccess() error {
	utils.MapEnvs(map[string]string{
		"CI_REGISTRY":          "DOCKER_HOST",
		"CI_REGISTRY_USER":     "DOCKER_USER",
		"CI_REGISTRY_PASSWORD": "DOCKER_PASSWORD",
	})

	if err := envconfig.Process("DOCKER", auth); err != nil {
		return err
	}

	// 1) Get path to Docker config
	dockerConfig := fmt.Sprintf("%s/.docker/config.json", home())

	// 2) Add item to auth
	authFile := dockerAuthFile{make(map[string]dockerAuthItem)}
	authStr := fmt.Sprintf("%s:%s", auth.User, auth.Password)
	authBase64 := base64.StdEncoding.EncodeToString([]byte(authStr))
	authFile.Auths[auth.Host] = dockerAuthItem{auth.User, auth.Password, authBase64}

	// 3) Store into config file
	// 3.1) Make sure, directory structure is correct
	if err := os.MkdirAll(fmt.Sprintf("%s/.docker", home()), 0755); err != nil {
		return fmt.Errorf("Cannot initialize docker config.json directory: %s", err)
	}

	// 3.2) Store file
	marshalled, _ := json.Marshal(&authFile)
	if err := ioutil.WriteFile(dockerConfig, marshalled, 0644); err != nil {
		return fmt.Errorf("Cannot write docker config.json: %s", err)
	}

	return nil
}
