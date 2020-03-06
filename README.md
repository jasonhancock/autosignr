# autosignr

[![Build Status](https://travis-ci.org/jasonhancock/autosignr.svg?branch=master)](https://travis-ci.org/jasonhancock/autosignr)
[![Go Report Card](https://goreportcard.com/badge/jasonhancock/autosignr)](https://goreportcard.com/report/jasonhancock/autosignr)

A [Custom Policy Executable](https://docs.puppetlabs.com/puppet/latest/reference/ssl_autosign.html#policy-based-autosigning) or a daemon (depending on your desired mode of operation) that watches for new Puppet CSRs from instances in AWS (or other clouds), validates the instances belong to you via the cloud provider's API, then signs the cert. Currently only supports AWS, Azure, GPC, but looking to add support for Openstack, Cloudstack, and generic REST APIs.

Autosignr can optionally be configured to validate pre-shared-keys embedded within the CSR. See the blurb in the Puppet Client Configuration section below for more details on how to embed a PSK into your CSRs.

**This is a work-in-progress and is far from complete. Use at your own risk**

## Puppet Client Configuration (optional)

The Puppet client can use the instance ID as the certname in puppet.conf.

```
# Configures an AWS instance to use the instance ID as the certname. Use something like this to set it before the puppet client starts for the first time:

CONF=/etc/puppetlabs/puppet/puppet.conf

INSTANCE_ID=`wget -q -O - http://169.254.169.254/latest/meta-data/instance-id`

grep certname $CONF > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "certname = $INSTANCE_ID" >> $CONF
else
    sed -i "s/certname =.*/certname = $INSTANCE_ID/" $CONF
fi
```

### Optionally embedding a pre-shared key in the CSR

To embed a PSK in your CSR, create the `csr_attributes.yaml` file in Puppet's `$confdir`. The file looks something like this:

```
extension_requests:
    pp_preshared_key: my_preshared_key
```

For the full list of possible extensions and options that can be put into that file, see the [Puppet documentation](https://docs.puppetlabs.com/puppet/latest/reference/ssl_attributes_extensions.html).

## Modes of operation

Autosignr can function as either a custom policy executable (a process run by the Puppetmaster when a new CSR arrives) or as a standalone daemon that watches for new CSR's and signs them.

### Configuring to run as a custom policy executable

To run as a custom policy executable, add the following to your Puppetmaster's puppet.conf in the `[master]` section:

```
autosign = /path/to/autosignr
```

### Running as a daemon

To run as a daemon, simply start the executable. It doesn't fork itself. The redhat packaging includes a systemd unit file which should allow you to start the service as:

```
service autosignr start
```

Or more correctly:

```
systemctl start autosignr.service
```

If you're on el6 or older, you'll have to craft an init script, run under a process supervisor, etc.

## Configuration Options

The configuration file lives at `/etc/autosignr/config.yaml`.

| Name            | Type                        | Description |
| --------------- | --------------------------- | ----------- |
| dir             | string                      | The path to the directory storing CSR files on the Puppetmaster |
| cmd\_sign       | string                      | The command to execute to sign valid certificates. Should contain `%s` that will be replaced with the cert name |
| logfile         | string                      | Optional. If specified, log to this file instead of STDOUT |
| accounts\_aws   | array of AWS account hashes | See AWS Account details below. Accounts are searched in the order specified. |
| accounts\_azure | array of Azure account hashes | See Azure Account details below.  Accounts are searched in the order specified. |
| accounts\_gcp   | array of GCP account hashes | See GPC Account details below.  Accounts are searched in the order specified. |
| preshared\_keys | array of strings            | Optional. Array of valid preshared keys embedded within the certificate's extension fields. See the Puppetlab's documentation on [SSL cert extensions](https://docs.puppetlabs.com/puppet/latest/reference/ssl_attributes_extensions.html) for more details on how to embed a PSK into your CSR's. If PSKs are defined in the configuration, then all CSR's will be required to have valid PSK's to be automatically signed |

Example Configuration:

```
dir: /etc/puppetlabs/puppet/ssl/ca/requests
cmd_sign: /opt/puppetlabs/bin/puppet cert sign %s
logfile: /var/log/autosignr/autosignr.log
accounts_aws:
  - name: jhancock aws packer
    key: AWS_KEY_ID
    secret: AWS_SECRET_KEY
    regions:
      - us-west-2
      - us-east-1
preshared_keys:
  - abc123
  - defghi
```

### Account Configuration

Each account must have a `name` specified. In addition, each type of account may require its own set of attributes.

| Name | Type    | Description |
| ---- | ------- | ----------- |
| name | string  | A short, descriptive name for this account. |

#### Account Type: aws

| Name      | Type             | Description |
| --------- | ---------------- | ----------- |
| key       | string           | AWS Key Id |
| secret    | string           | AWS Secret Key |
| regions   | array of strings | A list of regions to check for each instance. Regions are searched in the order specified |
| attribute | string           | Optional. Defaults to `instance-id`. The name of the attribute to compare the certname against. Useful for using a tag to compare against instead of the instance id...set to `tag:Name` to compare against the `Name` tag |      

#### Account Type: azure

| Name            | Type             | Description |
| --------------- | ---------------- | ----------- |
| client_id       | string           | Azure Client Id |
| client_secret   | string           | Azure Secret Key |
| subscription_id | string           | Subscription Id |
| tenant_id       | string           | Tenant Id |
| attribute       | string           | Optional.  Defaults comparing the certname with Tags `Name`.  If you want to use another Tag specify the key value here |      
| dns_zone        | string           | Optional.  This will compare the certname with the computer name plus the Tag `dns_zone`.  This option is useful for Scale Sets | 

#### Account Type: gcp

| Name            | Type             | Description |
| --------------- | ---------------- | ----------- |
| project_id      | string           | project_id |
| credential_file | string           | The path to the key json |

GCP checks the hostnames on instances with the cert name.  Other attributes not supported at this time.

---
# autocleanr

Monitor the PuppetDB for inactive nodes (report_timestamp < X hours) then use the cloud provider's API to verify the certname no longer exists.  Then runs custom commands to deactivate and clean the cert name.

### Run as a CronJob
Example

```
15 */4 * * * /usr/sbin/autocleanr > /dev/null 2>&1
```

### Configuration Options

The configuration file lives at ` /etc/autosignr/autocleanr.yaml`.

| Name             | Type                        | Description |
| ---------------  | --------------------------- | ----------- |
| logfile          | string                      | Optional. If specified, log to this file instead of STDOUT |
| clean\_commands  | array of string             | The command sto execute to deactivate and clean node from puppet. Should contain `%s` that will be replaced with the cert name |
| include\_filters | array of strings            | Optional.  Additional PQL filters to add to the query to find inactive nodes.  More information in below setting
| inactive\_hours  | int                         | Number of hours before the node is considered inactive. |
| puppetdb_host    | string                      | The Hostname for the PuppetDB node |
| puppetdb\_protocol | string                    | The protocol to connect to puppetdb (http|https) |
| uppetdb\_ignore\_cert\_errors | boolean        | Set to true if you want to ignore any cert errors.  Should only be set to true in development environments |
| puppetdb\_nodes\_uri | string                  |  The Root endpoint for the PuppetDB. |


Example Configuration:

```
logfile: /var/log/autosignr/autocleanr.log
clean_commands:
  - /opt/puppetlabs/bin/puppet node deactivate %s
  - /opt/puppetlabs/bin/puppet node clean %s
include_filters:
  - and facts{name = \"locations\" and value in [\"aws\"]}
inactive_hours: 4
puppetdb_host: puppetdb.example.com
puppetdb_protocol: https
puppetdb_ignore_cert_errors: false
puppetdb_nodes_uri: /api/pdb/query/v4
```

#### include\_filters
Optional setting to include additional queries to the API call to PuppetDB.  

This setting currently supports [PQL](https://puppet.com/docs/puppetdb/5.2/api/query/v4/pql.html)  

If include_filters is omitted:
```
curl -XPOST -H 'Content-Type:application/json' "http://localhost:8080/pdb/query/v4" \
-d '{ "query": "nodes[certname]{ report_timestamp < \"2019-01-26T04:02:27Z\" }" }'
```

If you add the following to the config file  
```
include_filters:
   - and facts{name = \"location\" and value in [\"aws\"]}
```

Adding the filter above results:
```
curl -XPOST -H 'Content-Type:application/json' "http://localhost:8080/pdb/query/v4" \
-d '{ "query": "nodes[certname]{ report_timestamp < \"2019-01-28T04:02:27Z\" and facts{name = \"role\" and value in [\"nomad-client\"]} }" }'
```
