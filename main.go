package main

import(
    "log"
    "os"
    "time"
    
    "github.com/rjeczalik/notify"
)

var (
        inAws  bool   = false
        inGce  bool   = false
        bucket string = ""
        region string = ""
    )

const (
    BUCKET_ENV_VAR = "BUCKET_NAME"
    REGION_ENV_VAR = "REGION"
)

func main() {
    bucket = os.Getenv(BUCKET_ENV_VAR)
    if bucket == "" {
        log.Fatal("Unable to determine bucket name. Make sure BUCKET_NAME environment variable is set.")
    }
    log.Println("Target bucket: ", bucket)

    region = os.Getenv(REGION_ENV_VAR)
    if region == "" {
        // default to us-east1 for amazon. google doesn't use the region
        region = "us-east1"
    }
    log.Println("Target region: ", region)

    // determine which cloud we are in
    channel := make(chan *CommonMetadata, 1)
    go introspectGCE(channel)
    go introspectAWS(channel)

    select {
    case cmdd := <-channel:
        if cmdd.cloud == "aws" {
            inAws = true
        } else if cmdd.cloud == "gce" {
            inGce = true
        } else {
            log.Fatal("Unsupported cloud provider. Currently only check for GCE and AWS.")
        }
    case <-time.After(time.Second * 1):
        log.Println("Unable to determine cloud provider. Currently only check for GCE and AWS.")
        // log.Fatal("Unable to determine cloud provider. Currently only check for GCE and AWS.")
    }


    // setup watcher to begin watching inotify system events
    dirsToWatch := []string{"/tmp/..."}
    events := []notify.Event{ notify.InMovedTo }
    watchChan := make(chan notify.EventInfo, 1)
    setupWatcher(watchChan, dirsToWatch, events)
    defer notify.Stop(watchChan)

    for {
        evt := <-watchChan
        log.Println("Event: ", evt)
    }
}

/**
 * Sets up a notify watcher to watch specified files.
 * @param { chan<- *notify.EventInfo } - channel you wish to listen on
 * @param { []string } - list of directories you wish to watch
 * @param { []Event  } - list of events you wish to listen on
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
