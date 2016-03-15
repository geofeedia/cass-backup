package main

import (
	"log"
	"os"
)

const (
    REGION_ENV_VAR = "REGION"
    BUCKET_ENV_VAR = "BUCKET_NAME"
)

/**
 * Returns the region we should be operating.
 * Either from environment variable or defaults to `us-east-1`
 * @return { string }
 */
func getRegion() string {
	region := os.Getenv(REGION_ENV_VAR)
    if region == "" {
        // default to us-east-1 for amazon. google doesn't use the region
        region = "us-east-1"
        log.Println("No region provided via environment variable. Default to using 'us-east-1'")
    }
    log.Println("Target region: ", region)
    return region
}

/**
 * Returns the bucket we should be uploading to.
 * Reads from environment variable BUCKET_NAME.
 * Fails hard if env var is not set.
 * @return { string }
 */
func getBucket() string {
	bucket := os.Getenv(BUCKET_ENV_VAR)
    if bucket == "" {
        log.Fatal("Unable to determine bucket name. Make sure BUCKET_NAME environment variable is set.")
    }
    log.Println("Target bucket: ", bucket)
    return bucket
}