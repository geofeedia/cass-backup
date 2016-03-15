# Cass-Backup

### Backup tool for Cassandra

Inspired heavily by https://github.com/JeremyGrosser/tablesnap.

* ~~Tool should automatically detect which cloud it is running in (AWS/GCE)~~
* Upload to GCS or S3 depending on cloud
* ~~Bucket name as configuration option passed by env var~~
* ~~Use IAM creds to get permissions on the upload bucket~~
* ~~Listen for FS events for triggering uploads (inotify on Linux)~~
* Use cloud SDK to support multi-part uploads for reliability of uploading large files
* Should upload full and incremental snapshot files

Currently only supports Linux inotify events.

Default event listened to is `syscall.IN_MOVED_TO`

The key will be in the form of `<machine_hostname>-<instance_id>/path/to/upload/file`

#### Expected environment variables
```no-highlight
BUCKET_NAME=some_bucket    # assumes the bucket already exists and does not currently create it if not.
REGION=us-east1   					 # only used for amazon, ignored for google. defaults to 'us-east-1'
```
