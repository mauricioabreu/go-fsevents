package main

import (
	"log"
	"os"
	"path"

	fsevents "github.com/tywkeene/go-fsevents"
)

func handleEvents(watcher *fsevents.Watcher) {
	watcher.StartAll()
	go watcher.Watch()
	log.Println("Waiting for events...")
	for {
		select {
		case event := <-watcher.Events:
			log.Printf("Descriptors list:", watcher.ListDescriptors())
			log.Printf("Event Name: %s Event Path: %s", event.Name, event.Path)

			// Root watch directory was deleted, panic
			log.Printf("Watcher %s event %s", watcher.RootPath, event.Path)
			if event.IsRootDeletion(watcher.RootPath) == true {
				panic("Root watch directory deleted!")
			}

			// Directory events
			if event.IsDirCreated() == true {
				log.Println("Directory created:", path.Clean(event.Path))
				if err := watcher.AddDescriptor(path.Clean(event.Path), 0); err != nil {
					panic(err)
				} else {
					// When a new dir is created, we need to add it to the descriptors list and start it
					descriptor := watcher.GetDescriptorByPath(path.Clean(event.Path))
					descriptor.Start(watcher.FileDescriptor)
					log.Printf("Watch descriptor created and started for %s", event.Path)
				}
			}

			if event.IsDirRemoved() == true {
				log.Println("Directory removed:", path.Clean(event.Path))
				if err := watcher.RemoveDescriptor(path.Clean(event.Path)); err != nil {
					panic(err)
				} else {
					log.Printf("Watch descriptor removed for %s", event.Path)
				}
			}

			if event.IsDirChanged() == true {
				log.Println("Directory changed: ", event.Name)
			}

			// File events
			if event.IsFileCreated() == true {
				log.Println("File created: ", event.Name)
			}
			if event.IsFileRemoved() == true {
				log.Println("File removed: ", event.Name)
			}
			if event.IsFileChanged() == true {
				log.Println("File changed: ", event.Name)
			}
			break
		case err := <-watcher.Errors:
			log.Println(err)
			break
		}
	}
}

func main() {
	if len(os.Args) == 1 {
		panic("Must specify directory to watch")
	}
	watchDir := os.Args[1]

	options := &fsevents.WatcherOptions{
		// Recursive flag will make a watcher recursive,
		// meaning it will go all the way down the directory
		// tree, and add descriptors for all directories it finds
		Recursive: true,
		// UseWatcherFlags will use the flag passed to NewWatcher()
		// for all subsequently created watch descriptors
		UseWatcherFlags: true,
	}

	// You might need to play with these flags to get the events you want
	// You can use these pre-defined flags that are declared in fsevents.go,
	// or the original inotify flags declared in the golang.org/x/sys/unix package
	inotifyFlags := fsevents.DirCreatedEvent | fsevents.DirRemovedEvent |
		fsevents.FileCreatedEvent | fsevents.FileRemovedEvent |
		fsevents.FileChangedEvent | fsevents.RootEvent

	w, err := fsevents.NewWatcher(watchDir, inotifyFlags, options)
	if err != nil {
		panic(err)
	}
	handleEvents(w)
}
