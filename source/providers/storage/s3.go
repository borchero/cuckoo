package storage

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	aws3 "github.com/aws/aws-sdk-go/service/s3"
	"go.borchero.com/typewriter"
)

type s3 struct {
	client *aws3.S3
	bucket string
	logger typewriter.CLILogger
}

// NewS3 configures a new AWS S3 instance for usage.
func NewS3(bucket string, logger typewriter.CLILogger) (Provider, error) {
	// 1) Make session
	if err := os.Setenv("AWS_SDK_LOAD_CONFIG", "1"); err != nil {
		return nil, fmt.Errorf("Failed setting required environment variable: %s", err)
	}

	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Failed creating session: %s", err)
	}

	// 2) Get client
	client := aws3.New(sess)

	return &s3{
		client: client,
		bucket: bucket,
		logger: logger,
	}, nil
}

func (s *s3) Upload(objects ...TransferObject) error {
	for _, object := range objects {
		if err := s.uploadObject(object); err != nil {
			return fmt.Errorf("Failed uploading to S3 bucket '%s': %s", s.bucket, err)
		}
	}
	return nil
}

func (s *s3) Delete(objects ...string) error {
	for _, object := range objects {
		if err := s.deleteObject(object); err != nil {
			return fmt.Errorf("Failed deleting from S3 bucket '%s': %s", s.bucket, err)
		}
	}
	return nil
}

func (s *s3) List() ([]string, error) {
	// 1) Make request
	params := &aws3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	}
	out, err := s.client.ListObjectsV2(params)
	if err != nil {
		return nil, fmt.Errorf("Failed to list objects: %s", err)
	}

	// 2) Get object names
	result := make([]string, len(out.Contents))
	for i, object := range out.Contents {
		result[i] = *object.Key
	}

	return result, nil
}

func (s *s3) uploadObject(object TransferObject) error {
	s.logger.Infof(
		"Uploading to S3 bucket '%s': %s => %s",
		s.bucket, object.LocalPath, object.BucketPath,
	)

	// 1) Open local file
	file, err := os.Open(object.LocalPath)
	if err != nil {
		return fmt.Errorf("Failed opening local file '%s': %s", object.LocalPath, err)
	}
	defer file.Close()

	// 2) Upload
	// 2.1) Get Mime
	mimeType, err := getMimeType(object.LocalPath, file)
	if err != nil {
		return fmt.Errorf("Failed getting mime type of local file '%s': %s", object.LocalPath, err)
	}

	// 2.2) Get parameters
	params := &aws3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(object.BucketPath),
		Body:        file,
		ContentType: aws.String(mimeType),
	}

	// 2.3) Upload
	if _, err := s.client.PutObject(params); err != nil {
		return fmt.Errorf("Failed uploading file to path '%s': %s", object.BucketPath, err)
	}

	return nil
}

func (s *s3) deleteObject(object string) error {
	s.logger.Infof("Deleting from S3 bucket '%s': %s", s.bucket, object)

	// 1) Get parameters
	params := &aws3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(object),
	}

	// 2) Delete
	if _, err := s.client.DeleteObject(params); err != nil {
		return fmt.Errorf("Failed deleting file at path '%s': %s", object, err)
	}

	return nil
}
