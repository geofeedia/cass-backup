package main

import (
    "log"
    "time"
    "path/filepath"
    
    "github.com/rjeczalik/notify"
)

var (
    inAws  bool = false
    inGce  bool = false
    metaData    = new(CommonMetadata)
)

func main() {
    // determine which cloud we are in
    channel := make(chan *CommonMetadata, 1)
    go introspectGCE(channel)
    go introspectAWS(channel)

    select {
    case cmdd := <-channel:
        if cmdd.cloud == "aws" {
            inAws = true
            metaData = cmdd
        } else if cmdd.cloud == "gce" {
            inGce = true
            metaData = cmdd
        } else {
            log.Fatal("Unsupported cloud provider. Currently only checking for GCE and AWS.")
        }
    case <-time.After(time.Second * 1):
        log.Println("Unable to determine cloud provider. Currently only checking for GCE and AWS.")
        // log.Fatal("Unable to determine cloud provider. Currently only check for GCE and AWS.")
    }

    // setup watcher to begin watching inotify system events
    // the ... allows for recursive subdirectories
    dirsToWatch := []string{"/data/..."}
    events := []notify.Event{ notify.InMovedTo }
    watchChan := make(chan notify.EventInfo, 1)
    setupWatcher(watchChan, dirsToWatch, events)
    defer notify.Stop(watchChan)

    for {
        evt := <-watchChan

        // only upload *.db files
        if filepath.Ext(evt.Path()) != ".db" { 
            continue
        }

        if inAws == true {
            uploadToS3(evt.Path(), metaData)
        } else if inGce == true {
            uploadToGcs(evt.Path(), metaData)
        } else {
            log.Fatal("Unrecognized cloud. Did not upload to either GCS or S3.")
        }

        log.Println("Path: ", evt.Path())
        log.Println("File: ", filepath.Base(evt.Path()))
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
