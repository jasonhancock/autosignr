package autosignr

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/fsnotify.v1"
)

func WatchDir(conf *Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					result, _ := CheckCert(conf, event.Name)
					if result {
						SignCert(conf, CertnameFromFilename(event.Name))
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	log.Printf("watching %s", conf.Dir)
	err = watcher.Add(conf.Dir)
	if err != nil {
		return err
	}
	<-done
	return nil
}
