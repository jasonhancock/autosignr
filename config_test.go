package autosignr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var data1 = `
dir: /etc/puppetlabs/puppet/ssl/ca/requests
cmd_sign: puppet cert sign %s
accounts_aws:
  - name: jhancock aws packer
    key: abc123
    secret: def456
    attribute: 'tag:Name'
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
	conf1, err := ParseYaml([]byte(data1))
	require.NoError(t, err)
	require.False(t, conf1.CheckPSK)

	require.Len(t, conf1.Accounts, 1)
	acct, ok := conf1.Accounts[0].(*AccountAWS)
	require.True(t, ok)
	require.Equal(t, "abc123", acct.Key)
	require.Equal(t, "def456", acct.Secret)
	require.Len(t, acct.Regions, 2)
	require.Equal(t, "us-west-2", acct.Regions[0])
	require.Equal(t, "tag:Name", acct.Attribute)

	conf2, err := ParseYaml([]byte(data2))
	require.NoError(t, err)
	require.True(t, conf2.CheckPSK)

	_, ok = conf2.PresharedKeys["abc123"]
	require.True(t, ok)

	_, ok = conf2.PresharedKeys["abc1234"]
	require.False(t, ok)
}
