package main

import (
	"log"
	"os"

	"golang.org/x/net/context"
    "golang.org/x/oauth2/google"	
	storage "google.golang.org/api/storage/v1"
)

const (
	scope = storage.DevstorageFullControlScope
)

/**
 * Uploads the file at the specified path to a GCS bucket as specified 
 * by the environment variable BUCKET_NAME
 * 
 * @param  { string } filePath - The path to the file to upload
 * @param  { *CommonMetadata } metaData - The instance metadata
 */
func uploadToGcs(filePath string, metaData *CommonMetadata) {
	var bucket = getBucket()

	client, err := google.DefaultClient(context.Background(), scope)
	if err != nil {
		log.Fatal("Unable to establish default client for GCE.")
	}

	service, err := storage.New(client)
	if err != nil {
		log.Fatal("Unable to create storage service.")
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error reading file at: ", filePath)
		log.Fatal(err)
	}

	// sanitize path to remove initial `/data` from filepath for upload
    sanitizedPath := filePath[5:len(filePath)]

    // object name is in the format of: <machine_hostname>-<instance_id>/path/to/upload/file
    obj := &storage.Object{
    	Name: metaData.hostname + "-" + metaData.instance_id + sanitizedPath,
    }

    // SDK should use the Google service account for auth creds
	result, err := service.Objects.Insert(bucket, obj).Media(file).Do()

	if err != nil {
		log.Printf("Error encountered. Unable to upload to GCS... ", err)
	} else {
		log.Printf("Successfully uploaded object %v to GCS at location %v\n\n", result.Name, result.SelfLink)
	}
	file.Close()
}