package cdn

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awscloudfront "github.com/aws/aws-sdk-go/service/cloudfront"
	"helm.sh/helm/v3/pkg/time"
)

type cloudfront struct {
	client       *awscloudfront.CloudFront
	distribution string
}

// NewCloudfront creates a new CDN provider backed by a AWS Cloudfront distribution.
func NewCloudfront(distribution string) (Provider, error) {
	// 1) Make session
	if err := os.Setenv("AWS_SDK_LOAD_CONFIG", "1"); err != nil {
		return nil, fmt.Errorf("Failed setting required environment variable: %s", err)
	}

	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed creating session: %s", err)
	}

	// 2) Get client
	client := awscloudfront.New(sess)

	return &cloudfront{
		client:       client,
		distribution: distribution,
	}, nil
}

func (c *cloudfront) Invalidate(path string) error {
	// 1) Create invalidation
	invalidation := &awscloudfront.CreateInvalidationInput{
		DistributionId: aws.String(c.distribution),
		InvalidationBatch: &awscloudfront.InvalidationBatch{
			CallerReference: aws.String(fmt.Sprintf("%d", time.Now().UnixNano())),
			Paths: &awscloudfront.Paths{
				Items: []*string{
					aws.String(path),
				},
				Quantity: aws.Int64(1),
			},
		},
	}

	// 2) Submit
	if _, err := c.client.CreateInvalidation(invalidation); err != nil {
		return fmt.Errorf(
			"Failed invalidation Cloudfront distribution '%s': %s", c.distribution, err,
		)
	}

	return nil
}
