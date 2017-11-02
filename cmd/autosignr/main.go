package main

import (
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/jasonhancock/autosignr"
	"github.com/pkg/errors"
)

var conf autosignr.Config

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	conf, err := autosignr.LoadConfigFile(autosignr.DefaultConfigFile)
	if err != nil {
		log.Fatal(errors.Wrap(err, "loading config file"))
	}

	if conf.LogFile != "" {
		f, err := os.OpenFile(conf.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
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
		err := autosignr.ExistingCerts(conf)
		if err != nil {
			log.Println(errors.Wrap(err, "existing certs"))
		}
		err = autosignr.WatchDir(conf)
		if err != nil {
			log.Fatal(err)
		}
	}
}
