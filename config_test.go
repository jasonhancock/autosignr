package autosignr

import (
	"testing"

	"github.com/cheekybits/is"
)

var data1 = `
dir: /etc/puppetlabs/puppet/ssl/ca/requests
cmd_sign: puppet cert sign %s
accounts_aws:
  - name: jhancock aws packer
    key: abc123
    secret: def456
    regions:
      - us-west-2
      - us-east-1
`

var data2 = `
dir: /etc/puppetlabs/puppet/ssl/ca/requests
cmd_sign: puppet cert sign %s
logfile: /tmp/logfile.log
accounts_aws:
  - name: jhancock aws packer
    key: abc123
    secret: def456
    regions:
      - us-west-2
      - us-east-1
preshared_keys:
  - abc123
  - def456
`

func TestConfigParsing(t *testing.T) {
	is := is.New(t)

	conf1, err := ParseYaml([]byte(data1))
	is.NoErr(err)
	is.False(conf1.CheckPSK)

	is.Equal(1, len(conf1.Accounts))
	acct, ok := conf1.Accounts[0].(*AccountAWS)
	is.OK(ok)
	is.Equal("abc123", acct.Key)
	is.Equal("def456", acct.Secret)
	is.Equal(2, len(acct.Regions))
	is.Equal("us-west-2", acct.Regions[0])

	conf2, err := ParseYaml([]byte(data2))
	is.NoErr(err)
	is.True(conf2.CheckPSK)

	_, ok = conf2.PresharedKeys["abc123"]
	is.OK(ok)

	_, ok = conf2.PresharedKeys["abc1234"]
	is.OK(!ok)
}
