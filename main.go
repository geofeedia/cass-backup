package main

import(
    "log"
    "time"
    
    "golang.org/x/exp/inotify"
)

func main() {

    channel := make(chan *CommonMetadata, 1)

    go introspectGCE(channel)
    go introspectAWS(channel)

    select {
    case cmdd := <-channel:
        log.Println(cmdd)
        log.Println("We're in the cloud")        
    case <-time.After(time.Second * 1):
        // Print an empty response if we can't determine the placement in a timely manner
        log.Println("We're running locally")
    }

    watcher, err := inotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    err = watcher.Watch("/tmp")
    if err != nil {
        log.Fatal(err)
    }

    for {
        select {
        case ev := <-watcher.Event:
            log.Println("event:", ev)
        case err := <-watcher.Error:
            log.Println("error:", err)
        }
    }
}
