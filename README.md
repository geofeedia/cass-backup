# Cass-Backup

### Backup tool for Cassandra

Inspired heavily by https://github.com/JeremyGrosser/tablesnap.

* Tool should automatically detect which cloud it is running in (AWS/GCE)
* Upload to GCS or S3 depending on cloud
* Bucket name as configuration option passed by env var
* Use IAM creds to get permissions on the upload bucket
* Listen for FS events for triggering uploads (inotify on Linux)
* Use cloud SDK to support multi-part uploads for reliability of uploading large files
* Should upload full and incremental snapshot files


#### Expected environment variables
```no-highlight
BUCKET_NAME='some_s3_or_gcs_bucket'
```
