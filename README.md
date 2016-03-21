# Cass-Backup

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
Initially we thought we could set up recursive watchers on the file tree rooted at `/data`, but
we realized there are unavoidable race conditions when doing this in that there is a possibility
that a subdirectory is created, which in turn triggers a watcher to be setup, but there could be
files and/or subdirectories within that subdirectory that trigger events prior to the watcher 
being completely initialized. Therefore we could not confidently say that we were listening on
**all** incoming file system events.

Our current implementation reverts to a less reactionary model and to a more polling type model.
Every 5 minutes and at boot we recursively scan (i.e. poll) from the root `/data` directory for any new
directory paths that include either `/data/**/snapshots/**` or `/data/**/backups/**`. Any paths found 
containing either of those directories is added immediately to a list of directories we want to watch and 
begin listening for file system events. To keep the number of watchers from growing too large we listen 
for delete events at which point we remove those paths from the watch list. Then as [`inotify`](http://man7.org/linux/man-pages/man7/inotify.7.html)
events that pass a certain set of criteria:

	1. It's either in the `/data/**/snapshots/` or `/data/**/backups/` directories
	2. It's not a directory

then we upload those files to a cloud bucket either S3 or GCS depending on which cloud you are currently
running the binary in.

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