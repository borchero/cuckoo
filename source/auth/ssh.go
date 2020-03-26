package auth

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type sshAuth struct {
	PrivateKey string `envconfig:"PRIVATE_KEY" required:"true"`
	KnownHosts string `envconfig:"KNOWN_HOSTS"`
}

// NewSSHAuth returns a component that is able to write SSH credentials such that the SSH agent can
// be used.
func NewSSHAuth() Component {
	return &sshAuth{}
}

func (*sshAuth) Name() string {
	return "SSH"
}

func (auth *sshAuth) EnsureAccess() error {
	if err := envconfig.Process("SSH", auth); err != nil {
		return err
	}

	// 1) Ensure .ssh directory
	sshDir := fmt.Sprintf("%s/.ssh", home())
	err := os.MkdirAll(sshDir, 0700)
	if err != nil {
		return errors.New("Cannot create ssh directory")
	}

	// 2) Write private key file
	privateKeyFile := fmt.Sprintf("%s/id", sshDir)
	privateKey := strings.ReplaceAll(auth.PrivateKey, "\\n", "\n")
	err = ioutil.WriteFile(privateKeyFile, []byte(privateKey), 0600)
	if err != nil {
		return errors.New("Cannot write private key")
	}

	// 3) Write known hosts
	if auth.KnownHosts != "" {
		knownHostsFile := fmt.Sprintf("%s/known_hosts", sshDir)
		knownHosts := strings.ReplaceAll(auth.KnownHosts, "\\n", "\n")
		err = ioutil.WriteFile(knownHostsFile, []byte(knownHosts), 0644)
		if err != nil {
			return errors.New("Cannot write known hosts")
		}
	}

	return nil
}
