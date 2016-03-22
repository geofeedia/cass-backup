package main

import (
    "log"
    "os"
    "strings"
    "path/filepath"
)

const (
    REGION_ENV_VAR = "REGION"
    BUCKET_ENV_VAR = "BUCKET_NAME"
)

/**
 * Returns the region we should be operating in.
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

/**
 * Determines if path is to a file or a directory.
 * If unable to locate path (i.e. it's been deleted, then we return true)
 * @param  { string } - the file path
 * @return {  bool  }
 */
func isDirectory(path string) bool {
    fileInfo, err := os.Stat(path)
    if err != nil {
        log.Printf("Unable to determine if path to file or directory exists: %s", path)
        return true
    }

    return fileInfo.IsDir()
}

/**
 * Determines if the path is a 'snapshots' or
 * 'backups' directory.
 * @param  { string } - the file path
 * @return {  bool  }
 */
func isSnapshotOrBackupDir(fpath string) bool {
    return (strings.Contains(fpath, "/snapshots") || strings.Contains(fpath, "/backups")) && isDirectory(fpath)
}

/**
 * Determines if path is under one of the `/data/system*` directories
 * @param { string } - the file path
 * @return { bool  }
 */
func isCassSystemDir(fpath string) bool {
    return strings.Contains(fpath, "/data/system")
}

/**
 * Determines if file has an extension of .db, is a file,
 * and is not the "temp.db" file since that breaks 
 * restoring the backups if it is present.
 * @param { string } - the file path
 * @return { bool  }
 */
func shouldUploadFile(fpath string) bool {
    return filepath.Ext(fpath) == ".db" &&
         !isDirectory(fpath) &&
         !strings.Contains(fpath, "/tmp.db")
}
