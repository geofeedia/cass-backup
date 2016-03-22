package main

import (
    "log"
    "time"
    "syscall"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/rjeczalik/notify"
)

const (
    DELETE_MASK      = 512
    SELF_DELETE_MASK = 1024 
)

var (
    metaData = new(CommonMetadata)
)

func main() {
    // determine which cloud we are in
    channel := make(chan *CommonMetadata, 1)
    go introspectGCE(channel)
    go introspectAWS(channel)

    select {
    case cmdd := <-channel:
        metaData = cmdd 
    case <-time.After(time.Second * 2):
        log.Fatal("Unable to determine cloud provider. Currently only checking for GCE and AWS.")
    }

    // setup watcher to begin watching inotify system events
    // the ... allows for recursive subdirectories
    var dirsToWatch = []string{"/data/..."}
    var events = []notify.Event{ 
        notify.InCreate,
        notify.InMovedTo,
        notify.InDelete,
        notify.InDeleteSelf,
    }

    // we use two channels here. 1 for the watch events
    // and 1 for the upload events which should only get
    // pumped with events that we know we want to upload.
    var watchChan  = make(chan notify.EventInfo, 10000)
    var uploadChan = make(chan notify.EventInfo, 10000)

    setupWatcher(watchChan, dirsToWatch, events)
    defer notify.Stop(watchChan)

    // start 2 concurrent upload workers to listen
    // on same channel
    go upload(uploadChan, metaData)
    go upload(uploadChan, metaData)

    for {
        evt := <-watchChan

        // if we receive a delete event, we need to stop that watcher 
        // and recreate watchers to avoid any leaks
        var mask = evt.Sys().(*syscall.InotifyEvent).Mask
        // syscall.IN_DELETE and syscall.IN_DELETE_SELF bit masks
        if  mask == DELETE_MASK || mask == SELF_DELETE_MASK {
            // stop removes any watchers
            notify.Stop(watchChan)
            setupWatcher(watchChan, dirsToWatch, events)
            continue
        }

        var fpath = evt.Path()

        // don't upload files in main table directory aka
        // working directory files.
        // we only want /data/keyspace/**/table/backups/* and /data/keyspace/**/table/snapshots/*
        // 
        // ignores any files under the /data/system* directories as well
        // 
        if isSnapshotOrBackupDir(fpath) == true && isCassSystemDir(fpath) == false {
            // pump events we want to upload to the 
            // upload channel
            uploadChan <- evt    
        }
    }
}

/**
 * Sets up a notify watcher to watch specified files.
 * @param { chan<- notify.EventInfo } - channel you wish to listen on
 * @param { []string }                - list of directories you wish to watch
 * @param { []Event  }                - list of events you wish to listen on
 */
func setupWatcher(channel chan<- notify.EventInfo, dirsToWatch []string, events []notify.Event) {
    for i := 0; i < len(dirsToWatch); i++ {
        if len(events) == 0 {
            log.Println("No listen events provided. No watchers set up.")
        }

        // loop over each event and add
        // it to the watcher
        for j := 0; j < len(events); j++ {
            err := notify.Watch(dirsToWatch[i], channel, events[j])
            if err != nil {
                var errStr = err.Error()
                // specific error message from github.com/rjeczalik/notify library
                // that we are choosing to ignore since it doesn't break anything
                if strings.Contains(errStr, "no such file or directory") {
                    log.Printf("No such file at %s to watch. Okay... moving on.", errStr)
                    break;
                // die on all other errors
                } else {
                    log.Println("Unable to watch directory at ", dirsToWatch[i])
                    log.Fatal(err)
                }
            }
        }
    }
}

/**
 * Handles upload to either S3 or GCS as indicated by available cloud metadata.
 * @param  { <-chan notify.EventInfo } - the upload channel we are listening on
 * @param  { *CommonMetadata         } - the cloud instance metadata
 */
func upload(channel <-chan notify.EventInfo, metaData *CommonMetadata) {
    var cloud = metaData.cloud
    for {
        // listen for events that we need to upload
        // and block when necessary
        evt := <-channel

        mask := evt.Sys().(*syscall.InotifyEvent).Mask
        // syscall.IN_DELETE and syscall.IN_DELETE_SELF bit masks
        if mask == DELETE_MASK || mask == SELF_DELETE_MASK {
            continue
        }

        var fpath = evt.Path()
        // if we have a directory then we need to upload
        // its contents
        if isDirectory(fpath) == true {
            // We do this to cover the race condition where a watcher is not
            // set up in time to catch any events from that directory.
            // There is a possibility that uploads will occur twice, but at the
            // moment that's not a problem since it's already so fast.
            // 
            // 3 seconds was chosen arbitrarily. No experimenting
            // was done for different times.
            time.Sleep(time.Second * 3)
            filepath.Walk(fpath, func(fpath string, f os.FileInfo, err error) error {
                // upload all files to either s3 or gcs
                if isDirectory(fpath) == false {
                    if hasDBExtension(fpath) == true {
                        uploadToCloud(fpath, cloud)
                    }
                }
                return nil
            })
        } else {
            uploadToCloud(fpath, cloud)
        }
    }
}

/**
 * actually makes the upload to the cloud
 * @param  { string } - path to file to upload
 * @param  { string } - the cloud, either "aws" or "gce"
 */
func uploadToCloud(fpath string, cloud string) {
    if cloud == "aws" {
        uploadToS3(fpath, metaData)
    } else if cloud == "gce" {
        uploadToGcs(fpath, metaData)
    } else {
        log.Fatal("Unsupported cloud provider. Currently only checking for GCE and AWS.")
    }
}
