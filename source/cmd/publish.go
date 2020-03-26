package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.borchero.com/cuckoo/providers/cdn"
	"go.borchero.com/cuckoo/providers/storage"
	"go.borchero.com/cuckoo/utils"
	"go.borchero.com/typewriter"
)

const publishDescription = `
The publish command can be used to push a directory containing files to some object storage bucket.
The intended use case for this command is to publish the contents of the bucket as a website.

Currently, the publish command enables pushing to an AWS S3 or Google Cloud Storage bucket.
Additionally, it enables invalidating an AWS Cloudfront cache if the AWS S3 bucket is served via
Cloudfront to enable CDN caching and TLS support.
`

type void struct{}

var publishArgs struct {
	dir             string
	provider        string
	bucket          string
	prefix          string
	purge           bool
	cloudfrontCache struct {
		name string
		path string
	}
}

func init() {
	publishCommand := &cobra.Command{
		Use:   "publish",
		Short: "Push a directory to an object storage bucket to publish as a website.",
		Long:  publishDescription,
		Args:  cobra.ExactArgs(0),
		Run:   runPublish,
	}

	publishCommand.Flags().StringVarP(
		&publishArgs.dir, "dir", "d", ".",
		"The directory from which to get files to upload.",
	)
	publishCommand.Flags().StringVarP(
		&publishArgs.bucket, "bucket", "b", "",
		"The bucket to which the files should be uploaded.",
	)

	publishCommand.Flags().StringVar(
		&publishArgs.provider, "provider", "s3",
		"The provider to use to access the bucket (s3/gcs).",
	)
	publishCommand.Flags().StringVar(
		&publishArgs.prefix, "prefix", "",
		"The prefix to use for storing paths in the bucket.",
	)

	publishCommand.Flags().BoolVar(
		&publishArgs.purge, "purge", false,
		"Whether to remove files from the bucket which are not present locally.",
	)

	publishCommand.Flags().StringVar(
		&publishArgs.cloudfrontCache.name, "aws-cloudfront-distribution", "",
		"The name of an AWS Cloudfront cache to invalidate after publishing.",
	)
	publishCommand.Flags().StringVar(
		&publishArgs.cloudfrontCache.path, "aws-cloudfront-path", "/*",
		"The paths to invalidate after publishing.",
	)

	rootCmd.AddCommand(publishCommand)
}

func runPublish(cmd *cobra.Command, args []string) {
	logger := typewriter.NewCLILogger()

	// 1) Verify parameters
	if publishArgs.dir == "" {
		typewriter.Fail(logger, "Source directory must be given", nil)
	}
	if publishArgs.bucket == "" {
		typewriter.Fail(logger, "Bucket must be given", nil)
	}

	// 2) Get storage provider
	var provider storage.Provider
	var err error = fmt.Errorf("Storage provider %s does not exist", publishArgs.provider)
	if publishArgs.provider == "s3" {
		provider, err = storage.NewS3(publishArgs.bucket, logger)
	} else if publishArgs.provider == "gcs" {
		provider, err = storage.NewGCS(context.Background(), publishArgs.bucket, logger)
	}
	if err != nil {
		typewriter.Fail(logger, "Failed to get storage provider", err)
	}

	// 3) Get objects to upload
	// 3.1) Iterate over directory
	files, err := utils.GetMatchingFiles(".*", publishArgs.dir)
	if err != nil {
		typewriter.Fail(logger, "Unable to get files of provided directory", err)
	}

	// 3.2) Get transfer objects
	prefixLen := len(publishArgs.dir)
	if strings.HasSuffix(publishArgs.dir, ".") {
		prefixLen--
	}
	transfer := make([]storage.TransferObject, len(files))
	bucketPaths := make(map[string]void)

	for i, file := range files {
		bucketPath := publishArgs.prefix + file[prefixLen:]
		transfer[i] = storage.TransferObject{
			LocalPath:  file,
			BucketPath: bucketPath,
		}
		bucketPaths[bucketPath] = void{}
	}

	// 4) Upload transfer objects
	if err := provider.Upload(transfer...); err != nil {
		typewriter.Fail(logger, "Failed to upload objects", err)
	}

	// 5) Invalidate cache if needed
	if publishArgs.cloudfrontCache.name != "" {
		logger.Infof("Creating invalidation...")

		// 5.1) Get provider
		cdnProvider, err := cdn.NewCloudfront(publishArgs.cloudfrontCache.name)
		if err != nil {
			typewriter.Fail(logger, "Failed to get CDN provider", err)
		}

		// 5.2) Invalidate
		if err := cdnProvider.Invalidate(publishArgs.cloudfrontCache.path); err != nil {
			typewriter.Fail(logger, "Failed to invalidate CDN paths", err)
		}
	}

	// 6) Purge old items if needed
	if publishArgs.purge {
		logger.Infof("Now purging objects...")

		// 6.1) List existing
		existing, err := provider.List()
		if err != nil {
			typewriter.Fail(logger, "Failed to list objects for purging", err)
		}

		// 6.2) Find the ones which are not present anymore
		removals := make([]string, 0)
		for _, path := range existing {
			if _, ok := bucketPaths[path]; !ok {
				removals = append(removals, path)
			}
		}

		if len(removals) == 0 {
			logger.Infof("Nothing to purge")
		}

		// 6.3) Remove
		if err := provider.Delete(removals...); err != nil {
			typewriter.Fail(logger, "Failed to purge objects", err)
		}
	}

	logger.Success("Done 🎉")
}
