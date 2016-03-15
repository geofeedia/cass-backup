package main

import (
	"log"
	"os"
	"bufio"
	"path/filepath"

    "github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
    REGION_ENV_VAR = "REGION"
    BUCKET_ENV_VAR = "BUCKET_NAME"
)

/**
 * Uploads the file at the specified path to an S3 bucket in a specific region
 * as indicated by the environment variables BUCKET_NAME and REGION
 * 
 * @param  { string } filePath - The path to the file to upload
 */
func uploadToS3(filePath string) {
	var bucket = os.Getenv(BUCKET_ENV_VAR)
    if bucket == "" {
        log.Fatal("Unable to determine bucket name. Make sure BUCKET_NAME environment variable is set.")
    }
    log.Println("Target bucket: ", bucket)

	var region = os.Getenv(REGION_ENV_VAR)
	region = os.Getenv(REGION_ENV_VAR)
    if region == "" {
        // default to us-east-1 for amazon. google doesn't use the region
        region = "us-east-1"
    }
    log.Println("Target region: ", region)

    f, fileIoErr := os.Open(filePath); 
    if fileIoErr != nil {
    	log.Printf("Error reading file at: ", filePath)
    	log.Fatal(fileIoErr)
    }

	uploader := s3manager.NewUploader(awssession.New(&aws.Config{Region: aws.String(region)}))
	result, err := uploader.Upload(&s3manager.UploadInput{	
		Body: bufio.NewReader(f),
		Bucket: aws.String(bucket),
		Key: aws.String(filepath.Base(filePath)),
	})

	f.Close()
	if err != nil {
		log.Println("Error encountered. Unable to upload to S3.")
		log.Fatal(err)
	}

	log.Println("Successfully uploaded to S3 at the following location: ", result.Location)
}