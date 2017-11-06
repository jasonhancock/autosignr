package main

import (
	"flag"
	"fmt"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/jasonhancock/autosignr"
)

var instanceRegex = regexp.MustCompile(`^i-\w+$`)

// Uses an autosignr configuration file to search multiple AWS accounts across
// multiple regions for an instance. If the instance is found, returns the
// specified tag if the instance has the tag.
func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	var (
		configFile = flag.String("config", autosignr.DefaultConfigFile, "Path to the autosignr config file")
		tag        = flag.String("tag", "Name", "Name of the tag to return")
	)

	flag.Parse()

	conf, err := autosignr.LoadConfigFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	if flag.NArg() != 1 {
		log.Fatal("invalid number of arguments")
	}

	instanceID := flag.Arg(0)

	if err := conf.Init(); err != nil {
		log.Fatal(err)
	}

	if !instanceRegex.MatchString(instanceID) {
		log.Fatalf("not an instance id: %s", instanceID)
	}

	for _, acct := range conf.Accounts {
		awsAcct, ok := acct.(*autosignr.AccountAWS)
		if !ok {
			continue
		}
		awsAcct.Init()

		result := awsAcct.GetInstance(instanceID)
		if result != nil {
			for _, v := range result.Tags {
				if *v.Key == *tag {
					fmt.Println(*v.Value)
				}
			}
		}
	}
}
