package autosignr

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Dir           string
	Cmdsign       string
	Logfile       string
	CheckPSK      bool
	Accounts      []Account
	Mycreds       []map[string]interface{}
	PresharedKeys map[string]bool
}

const DefaultConfigFile string = "/etc/autosignr/config.yaml"

func (f *Config) LoadConfigFile(filename string) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	f.ParseYaml(yamlFile)
}

func (f *Config) ParseYaml(yamldata []byte) {
	var data map[string]interface{}

	err := yaml.Unmarshal(yamldata, &data)
	if err != nil {
		panic(err)
	}

	f.Dir = data["dir"].(string)
	f.Cmdsign = data["cmd_sign"].(string)
	f.PresharedKeys = make(map[string]bool)

	if _, ok := data["logfile"]; ok {
		f.Logfile = data["logfile"].(string)
	}

	if _, ok := data["preshared_keys"]; ok {
		for _, e := range data["preshared_keys"].([]interface{}) {
			f.PresharedKeys[e.(string)] = true
			f.CheckPSK = true
		}
	}

	f.Accounts = make([]Account, len(data["credentials"].([]interface{})))

	for i, e := range data["credentials"].([]interface{}) {
		switch e.(map[interface{}]interface{})["type"].(string) {
		case "aws":
			f.Accounts[i] = NewAccountAWS(e.(map[interface{}]interface{}))
		default:
			panic(fmt.Sprintf("Unsupported Account type: %s", e.(map[interface{}]interface{})["type"].(string)))
		}
	}
}
