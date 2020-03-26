package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	gcloud "cloud.google.com/go/storage"
	"go.borchero.com/typewriter"
	"google.golang.org/api/iterator"
)

type gcs struct {
	client *gcloud.Client
	bucket string
	logger typewriter.CLILogger
}

// NewGCS configures a new Google Cloud Storage instance for usage.
func NewGCS(ctx context.Context, bucket string, logger typewriter.CLILogger) (Provider, error) {
	client, err := gcloud.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &gcs{
		client: client,
		bucket: bucket,
		logger: logger,
	}, nil
}

func (s *gcs) Upload(objects ...TransferObject) error {
	for _, object := range objects {
		if err := s.uploadObject(object); err != nil {
			return fmt.Errorf("Failed uploading to GCS bucket '%s': %s", s.bucket, err)
		}
	}
	return nil
}

func (s *gcs) Delete(objects ...string) error {
	for _, object := range objects {
		if err := s.deleteObject(object); err != nil {
			return fmt.Errorf("Failed deleting from GCS bucket '%s': %s", s.bucket, err)
		}
	}
	return nil
}

func (s *gcs) List() ([]string, error) {
	// 1) Make request
	ctx := context.Background()
	query := &gcloud.Query{Versions: false}
	it := s.client.Bucket(s.bucket).Objects(ctx, query)

	// 2) Get all objects
	result := make([]string, 0)
	for {
		attributes, err := it.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, fmt.Errorf("Failed listing objects: %s", err)
		}
		result = append(result, attributes.Name)
	}

	return result, nil
}

func (s *gcs) uploadObject(object TransferObject) error {
	s.logger.Infof(
		"Uploading to GCS bucket '%s': %s => %s",
		s.bucket, object.LocalPath, object.BucketPath,
	)

	// 1) Open local file
	file, err := os.Open(object.LocalPath)
	if err != nil {
		return fmt.Errorf("Failed opening local file '%s': %s", object.LocalPath, err)
	}
	defer file.Close()

	// 2) Get writer to bucket
	// 2.1) Get context (30 second write timeout)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	// 2.2) Get bucket writer
	writer := s.client.Bucket(s.bucket).Object(object.BucketPath).NewWriter(ctx)
	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("Failed uploading file to path '%s': %s", object.BucketPath, err)
	}

	// 2.3) Finalize
	if err := writer.Close(); err != nil {
		return fmt.Errorf("Failed finalizing upload to path '%s': %s", object.BucketPath, err)
	}

	return nil
}

func (s *gcs) deleteObject(object string) error {
	s.logger.Infof("Deleting from S3 bucket '%s': %s", s.bucket, object)

	ctx := context.Background()
	if err := s.client.Bucket(s.bucket).Object(object).Delete(ctx); err != nil {
		return fmt.Errorf("Failed deleting file at path '%s': %s", object, err)
	}

	return nil
}
