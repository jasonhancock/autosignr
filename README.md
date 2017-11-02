# autosignr

[![Build Status](https://travis-ci.org/jasonhancock/autosignr.svg?branch=master)](https://travis-ci.org/jasonhancock/autosignr)
[![Go Report Card](https://goreportcard.com/badge/jasonhancock/autosignr)](https://goreportcard.com/report/jasonhancock/autosignr)

A [Custom Policy Executable](https://docs.puppetlabs.com/puppet/latest/reference/ssl_autosign.html#policy-based-autosigning) or a daemon (depending on your desired mode of operation) that watches for new Puppet CSRs from instances in AWS (or other clouds), validates the instances belong to you via the cloud provider's API, then signs the cert. Currently only supports AWS, but looking to add support for Openstack, Cloudstack, and generic REST APIs.

Autosignr can optionally be configured to validate pre-shared-keys embedded within the CSR. See the blurb in the Puppet Client Configuration section below for more details on how to embed a PSK into your CSRs.

**This is a work-in-progress and is far from complete. Use at your own risk**

## Puppet Client Configuration

The Puppet client must use the instance ID as the certname in puppet.conf.

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

| Name    | Type             | Description |
| ------- | ---------------- | ----------- |
| key     | string           | AWS Key Id |
| secret  | string           | AWS Secret Key |
| regions | array of strings | A list of regions to check for each instance. Regions are searched in the order specified |
