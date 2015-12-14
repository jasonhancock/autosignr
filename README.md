# autosignr

A daemon that watches for new Puppet CSRs from instances in AWS (or other clouds), validates the instances belong to you via the cloud provider's API, then signs the cert. Currently only supports AWS, but looking to add support for Openstack, Cloudstack, and generic REST APIs.

Can optionally be configured to validate pre-shared-keys embedded within the CSR. See [SSL cert extensions](https://docs.puppetlabs.com/puppet/latest/reference/ssl_attributes_extensions.html) for more details on how to embed a PSK into your CSR's

**This is a work-in-progress and is far from complete. Use at your own risk**

## Configuration Options

| Name            | Type                    | Description |
| dir             | string                  | The path to the directory storing CSR files on the Puppetmaster |
| cmd\_sign       | string                  | The command to execute to sign valid certificates. Should contain `%s` that will be replaced with the cert name |
| logfile         | string                  | Optional. If specified, log to this file instead of STDOUT |
| credentials     | array of account hashes | Each account needs to have a `type` key specified. More details below. Accounts are searched in the order specified. |
| preshared\_keys | array of strings        | Optional. Array of valid preshared keys embedded within the certificate's extension fields. See the Puppetlab's documentation on [SSL cert extensions](https://docs.puppetlabs.com/puppet/latest/reference/ssl_attributes_extensions.html) for more details on how to embed a PSK into your CSR's. If PSKs are defined in the configuration, then all CSR's will be required to have valid PSK's to be automatically signed |

Example Configuration:

```
dir: /etc/puppetlabs/puppet/ssl/ca/requests
cmd_sign: /opt/puppetlabs/bin/puppet cert sign %s
logfile: /var/log/autosignr/autosignr.log
credentials:
  - name: jhancock aws packer
    type: aws
    key_id: AWS_KEY_ID
    secret_key: AWS_SECRET_KEY
    regions:
      - us-west-2
      - us-east-1
preshared_keys:
  - abc123
  - defghi
```

### Account Configuration

Each account must have a `name` and `type` specified. In addition, each type of account may require its own set of attributes.

| Name | Type    | Description |
| name | string  | A short, descriptive name for this account. |
| type | string  | The account type ("aws", etc.) |

#### Account Type: aws

| Name | Type    | Description |
| key\_id     | string           | AWS Key Id |
| secret\_key | string           | AWS Secret Key |
| regions     | array of strings | A list of regions to check for each instance. Regions are searched in the order specified |


# TODO:
* OpenStack Support
* CloudStack Support
* Random REST Support?
