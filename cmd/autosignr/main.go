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

	go autosignr.ExistingCerts(conf)
	autosignr.WatchDir(conf)
}
