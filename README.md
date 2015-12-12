# autosignr

A daemon that watches for new Puppet CSRs from instances in AWS (or other clouds), validates the instances belong to you via the cloud provider's API, then signs the cert. Currently only supports AWS, but looking to add support for Openstack, Cloudstack, and generic REST APIs.
