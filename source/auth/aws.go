package auth

import "github.com/kelseyhightower/envconfig"

type awsAuth struct {
	AccessKeyID     string `envconfig:"ACCESS_KEY_ID" required:"true"`
	SecretAccessKey string `envconfig:"SECRET_ACCESS_KEY" required:"true"`
	DefaultRegion   string `envconfig:"DEFAULT_REGION" required:"true"`
}

// NewAWSAuth returns a component able to check authentication for AWS.
func NewAWSAuth() Component {
	return &awsAuth{}
}

func (*awsAuth) Name() string {
	return "Amazon Web Services"
}

func (auth *awsAuth) EnsureAccess() error {
	if err := envconfig.Process("AWS", auth); err != nil {
		return err
	}
	return nil
}
