package main

import (
    "log"
    "time"
    "syscall"
    "path"
    "path/filepath"
    
    "github.com/rjeczalik/notify"
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
        notify.InCloseWrite,
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
        if  mask == 512 || mask == 1024 {
            // stop removes any watchers
            notify.Stop(watchChan)
            setupWatcher(watchChan, dirsToWatch, events)
            continue
        }

        var fpath = evt.Path()

        // only upload files
        if isDirectory(fpath) == true { 
            continue
        }

        // don't upload files in main table directory aka
        // working directory files. 
        // we only want /data/keyspace/.../table/backups and /data/keyspace/.../table/snapshots
        // 
        // filepath.Dir strips the file off the path and
        // path.Base returns the last directory
        lastDir := path.Base(filepath.Dir(fpath))
        if lastDir != "snapshot" || lastDir != "backups" {
            continue
        }

        // pump events we want to upload to the 
        // upload channel
        uploadChan <- evt
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
                log.Println("Unable to watch directory at ", dirsToWatch[i])
                log.Fatal(err)
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
    for {
        // listen for events that we need to upload
        // and block when necessary
        evt := <-channel
        if metaData.cloud == "aws" {
            uploadToS3(evt.Path(), metaData)
        } else if metaData.cloud == "gce" {
            uploadToGcs(evt.Path(), metaData)
        } else {
            log.Fatal("Unsupported cloud provider. Currently only checking for GCE and AWS.")
        }
    }
}
