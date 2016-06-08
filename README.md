# Cass-Backup

## DISCLAIMER
Unfortunately this does not currently work as expected. The use of `inotify` watchers cannot keep up with the
rate and volume of `inotify` events generated from running `nodetool snapshot` and `nodetool repair` for
Cassandra. This is due to the inherent race condition in watching `inotify` events in that a watcher cannot
be created in time before an event (think writing a table snapshot to disk) is fired.

### Backup tool for Cassandra

#### Credits
Inspired by https://github.com/JeremyGrosser/tablesnap

#### Requirements
* ~~Tool should automatically detect which cloud it is running in (AWS/GCE)~~
* ~~Upload to GCS or S3 depending on cloud~~
* ~~Bucket name as configuration option passed by env var~~
* ~~Use IAM creds to get permissions on the upload bucket~~
* ~~Listen for FS events for triggering uploads (inotify on Linux)~~
* ~~Use cloud SDK to support multi-part uploads for reliability of uploading large files~~
* ~~Should upload full and incremental snapshot files~~

#### Overview
Currently listens on `/data/**/` recursively for a subset of [`inotify`](http://man7.org/linux/man-pages/man7/inotify.7.html)
events and as we receive them if they pass a certain set of criteria:

	1. It's either in the `/data/**/snapshots/` or `/data/**/backups/` directories
	2. It's not a directory

we upload those files to a cloud bucket in either S3 or GCS depending on which cloud you are currently
operating in.

#### Caveats
Currently only supports Linux.

Currently listening on `syscall.IN_MOVED_TO, syscall.IN_DELETE, syscall.IN_DELETE_SELF, syscall.IN_CREATE` events.

The DELETE events are for removing watchers for when snapshots and backups are cleaned up.

The key will be in the form of `<machine_hostname>-<instance_id>/path/to/file/to/upload`

#### Expected environment variables
```no-highlight
BUCKET_NAME=some_bucket    # assumes the bucket already exists and does not currently create it if not.
REGION=us-east1   		   # only used for amazon, ignored for google. defaults to 'us-east-1'
```

To build you can use the Makefile or just use the `go install` command.

##### Makefile
```bash
# will output binary in project folder as `cass-backup`
$ make
...
$ ./cass-backup
```