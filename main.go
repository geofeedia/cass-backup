package main

import (
    "log"
    "time"
    "syscall"
    "path/filepath"
    "os"
    
    "github.com/rjeczalik/notify"
)

const ROOT_DATA_DIR = "/data"

var (
    metaData    = new(CommonMetadata)
    dirsToWatch = map[string]string{}
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

    // we use two channels here. 1 for the watch events
    // and 1 for the upload events which should only get
    // pumped with events that we know we want to upload.
    var watchChan  = make(chan notify.EventInfo, 10000)
    var uploadChan = make(chan notify.EventInfo, 10000)
    var events = []notify.Event{ 
        notify.InCreate,
        notify.InMovedTo,
        notify.InDelete,
        notify.InDeleteSelf,
    }

    defer notify.Stop(watchChan)
    defer notify.Stop(uploadChan)

    // start 2 concurrent upload workers to listen
    // on same channel
    go upload(uploadChan, metaData)
    go upload(uploadChan, metaData)
    
    // start listening for watch events
    go func() {
        for {
            evt := <-watchChan
            // if we receive a delete event, we need to stop that watcher 
            // and recreate watchers to avoid any leaks
            mask := evt.Sys().(*syscall.InotifyEvent).Mask
            
            // syscall.IN_DELETE and syscall.IN_DELETE_SELF bit masks
            if  mask == 512 || mask == 1024 {
                // stop removes any watchers
                notify.Stop(watchChan)
                // remove from dirs to watch
                delete(dirsToWatch, evt.Path())
                continue
            }

            // pump upload channel
            uploadChan <- evt
        }
    }()

    // every 5 minutes check for new /data/**/snapshots/ or /data/**/backups/
    for {
        updateWatchedFiles()
        updateWatchers(watchChan, events, metaData)
        time.Sleep(time.Second * 300)        
    } 
}

/**
 * function that will be called upon the walking of each file
 * in updateWatchedFiles()
 * @param { string      } - path to the file or directory in the filepath.Walk(...) func
 * @param { os.FileInfo } - info about the specific file
 * @param { error       } - the error passed via the filepath.Walk(...)
 */
func walkDirFunc(fpath string, f os.FileInfo, err error) error {
    if isSnapshotOrBackupDir(fpath) == true {
        // only add it if doesn't already exist
        // i.e. if the key,value pair doesn't exist
        // since requesting a key that doesn't exist
        // will return a zero value for the specific data
        // type which for strings is an empty zero 
        // length string
        if len(dirsToWatch[fpath]) == 0 {
            dirsToWatch[fpath] = fpath
        }
    }
    return nil
}

/**
 * updates directories to watch by recursively
 * looking down from a root directory in this case '/data'
 */
func updateWatchedFiles() {
    err := filepath.Walk(ROOT_DATA_DIR, walkDirFunc)
    if err != nil {
        log.Fatal("Error trying to walk directories at root: ", ROOT_DATA_DIR)
    }
}

/**
 * updates which directories should be watched
 * @param  { chan<- notify.EventInfo } - channel to watch events on
 * @param  { []notify.Event          } - events to watch
 * @param { *CommonMetadata          } - cloud metadata
 */
func updateWatchers(watchChan chan<- notify.EventInfo, events []notify.Event, metaData *CommonMetadata) {
    for _, value := range dirsToWatch {
        if len(events) == 0 {
            log.Println("No listen events provided. No watchers set up.")
        }

        // loop over each event and add to the watcher
        for i := 0; i < len(events); i++ {
            err := notify.Watch(value, watchChan, events[i])
            if err != nil {
                log.Println("Unable to watch directory at ", value)
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
    var cloud = metaData.cloud
    for {
        // listen for events that we need to upload
        // and block when necessary
        evt := <-channel

        var fpath = evt.Path()
        // if we have a directory then we need to upload
        // its contents
        if isDirectory(fpath) == true {
            // wait a little while to avoid missing any
            // files being created 
            time.Sleep(time.Second * 10)
            filepath.Walk(fpath, func(fpath string, f os.FileInfo, err error) error {
                // upload all files to either s3 or gcs
                if isDirectory(fpath) == false {
                    uploadToCloud(fpath, cloud)
                }
                return nil
            })
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



////////////////////////////////////////////////////////////////////////
// OLD CODE
// 
// 
// package main

// import (
//     "log"
//     "time"
//     "syscall"
    
//     "github.com/rjeczalik/notify"
// )

// var (
//     metaData = new(CommonMetadata)
// )

// func main() {
//     // determine which cloud we are in
//     channel := make(chan *CommonMetadata, 1)
//     go introspectGCE(channel)
//     go introspectAWS(channel)

//     select {
//     case cmdd := <-channel:
//         metaData = cmdd 
//     case <-time.After(time.Second * 2):
//         log.Fatal("Unable to determine cloud provider. Currently only checking for GCE and AWS.")
//     }

//     // setup watcher to begin watching inotify system events
//     // the ... allows for recursive subdirectories
//     var dirsToWatch = []string{"/data/..."}
//     var events = []notify.Event{ 
//         notify.InCreate,
//         notify.InMovedTo,
//         notify.InDelete,
//         notify.InDeleteSelf,
//     }

//     // we use two channels here. 1 for the watch events
//     // and 1 for the upload events which should only get
//     // pumped with events that we know we want to upload.
//     var watchChan  = make(chan notify.EventInfo, 10000)
//     var uploadChan = make(chan notify.EventInfo, 10000)

//     setupWatcher(watchChan, dirsToWatch, events)
//     defer notify.Stop(watchChan)

//     // start 2 concurrent upload workers to listen
//     // on same channel
//     go upload(uploadChan, metaData)
//     go upload(uploadChan, metaData)

//     for {
//         evt := <-watchChan

//         // if we receive a delete event, we need to stop that watcher 
//         // and recreate watchers to avoid any leaks
//         var mask = evt.Sys().(*syscall.InotifyEvent).Mask
//         // syscall.IN_DELETE and syscall.IN_DELETE_SELF bit masks
//         if  mask == 512 || mask == 1024 {
//             // stop removes any watchers
//             notify.Stop(watchChan)
//             setupWatcher(watchChan, dirsToWatch, events)
//             continue
//         }

//         var fpath = evt.Path()

//         // only upload files
//         if isDirectory(fpath) == true { 
//             continue
//         }

//         // don't upload files in main table directory aka
//         // working directory files.
//         // we only want /data/keyspace/**/table/backups/* and /data/keyspace/**/table/snapshots/*
//         // 
//         // filepath.Dir strips the file off the path and
//         // path.Base returns the last directory
//         if isInSnapshotOrBackup(fpath) == true {
//             // pump events we want to upload to the 
//             // upload channel
//             uploadChan <- evt    
//         }
//     }
// }

// /**
//  * Sets up a notify watcher to watch specified files.
//  * @param { chan<- notify.EventInfo } - channel you wish to listen on
//  * @param { []string }                - list of directories you wish to watch
//  * @param { []Event  }                - list of events you wish to listen on
//  */
// func setupWatcher(channel chan<- notify.EventInfo, dirsToWatch []string, events []notify.Event) {
//     for i := 0; i < len(dirsToWatch); i++ {
//         if len(events) == 0 {
//             log.Println("No listen events provided. No watchers set up.")
//         }

//         // loop over each event and add
//         // it to the watcher
//         for j := 0; j < len(events); j++ {
//             err := notify.Watch(dirsToWatch[i], channel, events[j])
//             if err != nil {
//                 log.Println("Unable to watch directory at ", dirsToWatch[i])
//                 log.Fatal(err)
//             }
//         }
//     }
// }

// /**
//  * Handles upload to either S3 or GCS as indicated by available cloud metadata.
//  * @param  { <-chan notify.EventInfo } - the upload channel we are listening on
//  * @param  { *CommonMetadata         } - the cloud instance metadata
//  */
// func upload(channel <-chan notify.EventInfo, metaData *CommonMetadata) {
//     for {
//         // listen for events that we need to upload
//         // and block when necessary
//         evt := <-channel
//         if metaData.cloud == "aws" {
//             uploadToS3(evt.Path(), metaData)
//         } else if metaData.cloud == "gce" {
//             uploadToGcs(evt.Path(), metaData)
//         } else {
//             log.Fatal("Unsupported cloud provider. Currently only checking for GCE and AWS.")
//         }
//     }
// }
