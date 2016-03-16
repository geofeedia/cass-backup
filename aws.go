package main

import (
	"log"
	"os"
	"bufio"

    "github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)


/**
 * Uploads the file at the specified path to an S3 bucket in a specific region
 * as indicated by the environment variables BUCKET_NAME and REGION
 * 
 * @param  { string          } filePath - The path to the file to upload
 * @param  { *CommonMetadata } metaData - The instance metadata
 */
func uploadToS3(filePath string, metaData *CommonMetadata) {
	var bucket = getBucket()
	var region = getRegion()

    file, err := os.Open(filePath); 
    if err != nil {
    	log.Printf("Error reading file at: ", filePath)
    	log.Fatal(err)
    }

    // sanitize path to remove initial `/data` from filepath for upload
    sanitizedPath := filePath[5:len(filePath)]

    // the SDK relies on IAM credentials
    // S3 upload manager uploads large file in smaller
    // parts and in parallel.
    // key is in format of: <machine_hostname>-<instance_id>/path/to/upload/file
	uploader := s3manager.NewUploader(awssession.New(&aws.Config{Region: aws.String(region)}))
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body: bufio.NewReader(file),
		Bucket: aws.String(bucket),
		Key: aws.String(metaData.hostname + "-" + metaData.instance_id + sanitizedPath),
	})

	if err != nil {
		log.Println("Error encountered. Unable to upload to S3... ", err)
	} else {
		log.Println("Successfully uploaded to S3 at the following location %v", result.Location)
	}
	file.Close()
}