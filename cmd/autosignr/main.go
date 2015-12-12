package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/jasonhancock/autosignr"
)

var conf autosignr.Config

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	conf.LoadConfigFile(os.Args[1])

	if conf.Logfile != "" {
		f, err := os.OpenFile(conf.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		log.SetOutput(f)
	}

	go autosignr.ExistingCerts(conf)
	autosignr.WatchDir(conf)
}
