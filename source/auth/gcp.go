package auth

import (
	"encoding/json"
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type gcpAuth struct {
	Credentials string `envconfig:"CREDENTIALS" required:"true"`
}

type gcpCredentials struct {
	Type                string `json:"type"`
	ProjectID           string `json:"project_id"`
	PrivateKeyID        string `json:"private_key_id"`
	PrivateKey          string `json:"private_key"`
	ClientEmail         string `json:"client_email"`
	ClientID            string `json:"client_id"`
	AuthURI             string `json:"auth_uri"`
	TokenURI            string `json:"token_uri"`
	AuthProviderCertURL string `json:"auth_provider_x509_cert_url"`
	ClientCertURL       string `json:"client_x509_cert_url"`
}

// NewGCPAuth returns a component able to authenticate client libraries against the Google Cloud
// Platform.
func NewGCPAuth() Component {
	return &gcpAuth{}
}

func (*gcpAuth) Name() string {
	return "Google Cloud Platform"
}

func (auth *gcpAuth) EnsureAccess() error {
	if err := envconfig.Process("GOOGLE", auth); err != nil {
		return err
	}

	var credentials gcpCredentials
	if err := json.Unmarshal([]byte(auth.Credentials), &credentials); err != nil {
		return fmt.Errorf("Failed to parse GOOGLE_CREDENTIALS: %s", err)
	}

	return nil
}
