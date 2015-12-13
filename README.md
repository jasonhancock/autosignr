# autosignr

A daemon that watches for new Puppet CSRs from instances in AWS (or other clouds), validates the instances belong to you via the cloud provider's API, then signs the cert. Currently only supports AWS, but looking to add support for Openstack, Cloudstack, and generic REST APIs.

**This is a work-in-progress and is far from complete. Use at your own risk**

# TODO:

* Support for PSK in [SSL cert extensions](https://docs.puppetlabs.com/puppet/latest/reference/ssl_attributes_extensions.html) - WIP
* OpenStack Support
* CloudStack Support
* Random REST Support?
