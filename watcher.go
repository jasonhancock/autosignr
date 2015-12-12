package autosignr

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

func WatchDir(conf Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					CheckCert(conf, event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	log.Printf("watching %s", conf.Dir)
	err = watcher.Add(conf.Dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
