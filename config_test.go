package autosignr_test

import (
	"testing"

	"github.com/jasonhancock/autosignr"
)

var data1 = `
dir: /etc/puppetlabs/puppet/ssl/ca/requests
cmd_sign: puppet cert sign %s
credentials:
  - name: jhancock aws packer
    type: aws
    key_id: abc123
    secret_key: def456
    regions:
      - us-west-2
      - us-east-1
`

var data2 = `
dir: /etc/puppetlabs/puppet/ssl/ca/requests
cmd_sign: puppet cert sign %s
logfile: /tmp/logfile.log
credentials:
  - name: jhancock aws packer
    type: aws
    key_id: abc123
    secret_key: def456
    regions:
      - us-west-2
      - us-east-1
preshared_keys:
  - abc123
  - def456
`

func TestConfigParsing(t *testing.T) {

	var conf1, conf2 autosignr.Config

	conf1.ParseYaml([]byte(data1))
	conf2.ParseYaml([]byte(data2))

	// Check PSK values
	if conf1.CheckPSK != false {
		t.Errorf("Expected conf1 to have CheckPSK set to false")
	}

	if conf2.CheckPSK != true {
		t.Errorf("Expected conf2 to have CheckPSK set to true")
	}

	// Verify the PSK is set in conf2
	if _, ok := conf2.PresharedKeys["abc123"]; !ok {
		t.Errorf("expected PSK abc123 not detected")
	}

	if _, ok := conf2.PresharedKeys["abc1234"]; ok {
		t.Errorf("unexpected PSK abc123i4 detected")
	}
}
