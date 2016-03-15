package main

import(
    "log"
    "os"
    "time"
    
    "golang.org/x/exp/inotify"
)

var (
        inAws  bool   = false
        inGce  bool   = false
        bucket string = ""
    )

const (
    BUCKET_ENV_VAR = "BUCKET_NAME"
)

func main() {
    bucket = os.Getenv(BUCKET_ENV_VAR)
    log.Println("Target bucket: ", bucket)

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
        // Print an empty response if we can't determine the placement in a timely manner
        log.Println("Unable to determine cloud provider. Currently only check for GCE and AWS.")
    }


    // setup watcher to begin watching inotify system events
    dirsToWatch := []string{"/tmp", "/home/parallels/Desktop"}
    watcher := SetupWatcher(dirsToWatch)

    for {
        select {
        case ev := <-watcher.Event:
            log.Println("event:", ev)
        case err := <-watcher.Error:
            log.Println("error:", err)
        }
    }
}

/**
 * Constructs an inotify watcher to watch specified files.
 * @return {*inotify.Watcher}
 */
func SetupWatcher(dirsToWatch []string) *inotify.Watcher {
    watcher, err := inotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    for i := 0; i < len(dirsToWatch); i++ {
        err := watcher.Watch(dirsToWatch[i])
        if err != nil {
            log.Println("Unable to watch directory at ", dirsToWatch[i])
            log.Fatal(err)
        }
    }
    return watcher
}
