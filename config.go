package autosignr

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const DefaultConfigFile string = "/etc/autosignr/config.yaml"

type Config struct {
	Dir           string
	CmdSign       string
	LogFile       string
	CheckPSK      bool
	PresharedKeys map[string]struct{}
	Accounts      []Account
}

type parsedConfig struct {
	Dir           string       `yaml:"dir"`
	CmdSign       string       `yaml:"cmd_sign"`
	LogFile       string       `yaml:"logfile"`
	AWSAccounts   []AccountAWS `yaml:"accounts_aws"`
	PresharedKeys []string     `yaml:"preshared_keys"`
}

func (c *Config) Init() error {
	for i := range c.Accounts {
		err := c.Accounts[i].Init()
		if err != nil {
			errors.Wrapf(err, "initializing account %s", c.Accounts[i])
		}
	}
	return nil
}

func (p *parsedConfig) Config() *Config {
	c := &Config{
		Dir:           p.Dir,
		CmdSign:       p.CmdSign,
		LogFile:       p.LogFile,
		Accounts:      make([]Account, 0, len(p.AWSAccounts)),
		PresharedKeys: make(map[string]struct{}, len(p.PresharedKeys)),
	}

	if len(p.PresharedKeys) > 0 {
		c.CheckPSK = true
	}

	for i := range p.PresharedKeys {
		c.PresharedKeys[p.PresharedKeys[i]] = struct{}{}
	}

	for i := range p.AWSAccounts {
		c.Accounts = append(c.Accounts, &p.AWSAccounts[i])
	}

	return c
}

func LoadConfigFile(filename string) (*Config, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}

	return ParseYaml(yamlFile)
}

func ParseYaml(yamldata []byte) (*Config, error) {
	var f parsedConfig

	err := yaml.Unmarshal(yamldata, &f)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling yaml")
	}

	return f.Config(), nil
}
