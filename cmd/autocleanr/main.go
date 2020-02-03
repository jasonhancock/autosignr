package main

import (
	"os"

	"github.com/jasonhancock/autosignr"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{})

	// autocleanr reads both the autosignr config file and the autocleanr specific config file
	conf, err := autosignr.LoadConfigFile(autosignr.DefaultConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	// Read in the autocleanr config
	viper.AutomaticEnv()
	viper.SetDefault("logfile", "/var/log/autosignr/autocleanr.log")
	viper.SetDefault("clean_commands", []string{})
	viper.SetDefault("include_facts", make(map[string][]string))
	viper.SetDefault("inactive_hours", 2)
	viper.SetDefault("puppetdb_host", "puppetdb")
	viper.SetDefault("puppetdb_protocol", "https")
	viper.SetDefault("puppetdb_ignore_cert_errors", false)
	viper.SetDefault("puppetdb_nodes_uri", "/api/pdb/query/v4/nodes")
	viper.SetConfigName("autocleanr")
	viper.AddConfigPath("/etc/autosignr")
	viper.AddConfigPath(".")
	err = viper.ReadInConfig()
	if err != nil {
		switch i := err.(type) {
		case viper.UnsupportedConfigError:
			log.Println("No config file")
		default:
			log.Fatalf("Fatal error config file: %s \n", i)
		}
	}

	// Wire up the log file to the logger
	f, err := os.OpenFile(viper.GetString("logfile"), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	if err := conf.Init(); err != nil {
		log.Fatal(err)
	}

	list, err := autosignr.FindInactiveNodes(
		viper.GetInt("inactive_hours"),
		viper.GetString("puppetdb_host"),
		viper.GetString("puppetdb_protocol"),
		viper.GetString("puppetdb_nodes_uri"),
		viper.GetBool("puppetdb_ignore_cert_errors"),
		viper.GetStringSlice("include_filters"))

	if err != nil {
		log.Fatal("Unable to retrieve node list: " + err.Error())
	}

	for _, certname := range list {
		result := false

		for _, acct := range conf.Accounts {
			if acct.Type() == "aws" {
				result = acct.Check(certname)
				if result {
					break
				}
			}
		}

		if !result {
			log.Println("Did not find the instance: " + certname)
			autosignr.CleanNode(viper.GetStringSlice("clean_commands"), certname)
		} else {
			log.Println("Found the instance: " + certname)
		}
	}
}
