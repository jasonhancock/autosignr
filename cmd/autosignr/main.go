package main

import (
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/jasonhancock/autosignr"
)

var conf autosignr.Config

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	conf.LoadConfigFile(autosignr.DefaultConfigFile)

	if conf.Logfile != "" {
		f, err := os.OpenFile(conf.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}
		defer f.Close()

		log.SetOutput(f)
	}

	// If we were passed an argument, that means we're operating as a custom
	// policy executable. Expect the certname as the only argument, and the
	// certificate data pem encoded on stdin
	if len(os.Args) > 1 {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}

		result, err := autosignr.ValidateCert(conf, data, os.Args[1])

		if err != nil {
			os.Exit(2)
		}

		if result {
			os.Exit(0)
		} else {
			os.Exit(1)
		}

	} else {
		// Operate in daemon mode, operate on fsnotify events
		go autosignr.ExistingCerts(conf)
		autosignr.WatchDir(conf)
	}
}
