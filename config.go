package autosignr

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Dir      string
	Cmdsign  string
	Logfile  string
	Accounts []Account
	Mycreds  []map[string]interface{}
}

func (f *Config) LoadConfigFile(filename string) {
	var data map[string]interface{}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &data)
	if err != nil {
		panic(err)
	}

	f.Dir = data["dir"].(string)
	f.Cmdsign = data["cmd_sign"].(string)

	if _, ok := data["logfile"]; ok {
		f.Logfile = data["logfile"].(string)
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
