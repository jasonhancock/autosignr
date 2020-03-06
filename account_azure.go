package autosignr

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	log "github.com/sirupsen/logrus"
)

var rgRegexp = regexp.MustCompile("resourceGroups/(.+?)/providers")

// AccountAzure encapsulates the account information for Azure
type AccountAzure struct {
	Name           string `yaml:"name"`
	Domain         string `yaml:"dns_zone"`
	ClientID       string `yaml:"client_id"`
	ClientSecret   string `yaml:"client_secret"`
	SubscriptionID string `yaml:"subscription_id"`
	TenantID       string `yaml:"tenant_id"`
	Attribute      string `yaml:"attribute"`

	vmClient     compute.VirtualMachinesClient
	vmssClient   compute.VirtualMachineScaleSetsClient
	vmssVMClient compute.VirtualMachineScaleSetVMsClient
}

// Init setup the account
func (a *AccountAzure) Init() error {
	if a.Attribute == "" {
		a.Attribute = "Name"
	}

	if a.Domain == "" {
		a.Domain = "dns_zone"
	}

	conf := auth.NewClientCredentialsConfig(a.ClientID, a.ClientSecret, a.TenantID)

	a.vmClient = compute.NewVirtualMachinesClient(a.SubscriptionID)
	a.vmssClient = compute.NewVirtualMachineScaleSetsClient(a.SubscriptionID)
	a.vmssVMClient = compute.NewVirtualMachineScaleSetVMsClient(a.SubscriptionID)

	var err error
	a.vmClient.Authorizer, err = conf.Authorizer()
	if err != nil {
		return err
	}

	a.vmssClient.Authorizer, err = conf.Authorizer()
	if err != nil {
		return err
	}

	a.vmssVMClient.Authorizer, err = conf.Authorizer()
	if err != nil {
		return err
	}
	return nil
}

// Type returns the type of account
func (a AccountAzure) Type() string {
	return "azure"
}

// Check look for the instanceID in the account
func (a *AccountAzure) Check(instanceID string) bool {
	log.WithFields(log.Fields{
		"instance":     instanceID,
		"subscription": a.Name,
	}).Debug("checking")

	// Check the Virtual Machines endpoints
	if vmCheck, err := a.checkVM(instanceID); err == nil && vmCheck {
		log.WithFields(log.Fields{
			"instance": instanceID,
			"account":  a.Name,
			"found":    true,
		}).Debug("check-result")
		return true
	}

	// Check the Virtual Machine Scale Sets
	if vmssCheck, err := a.checkScaleSetVM(instanceID); err == nil && vmssCheck {
		log.WithFields(log.Fields{
			"instance": instanceID,
			"account":  a.Name,
			"found":    true,
		}).Debug("check-result")
		return true
	}

	log.WithFields(log.Fields{
		"instance": instanceID,
		"account":  a.Name,
		"found":    false,
	}).Debug("check-result")
	return false
}

// String returns the account info
func (a AccountAzure) String() string {
	return fmt.Sprintf("azure account: %s", a.Name)
}

// checkVM will look for the instance in Azure Virtual Machines
func (a AccountAzure) checkVM(instanceID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Check the Virtual Machines endpoints
	for subList, err := a.vmClient.ListAll(ctx, ""); subList.NotDone(); err = subList.NextWithContext(ctx) {
		if err != nil {
			return false, err
		}

		for _, instance := range subList.Values() {
			// Check for the name tag
			if val, ok := instance.Tags[a.Attribute]; ok {
				if *val == instanceID {
					return true, nil
				}
			}

			// Check for the dns_zone tag
			if val, ok := instance.Tags[a.Domain]; ok {
				if fmt.Sprintf("%s.%s", *instance.OsProfile.ComputerName, *val) == instanceID {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// getAllScaleSets will pull all the Azure Virtual Machine Scale Set in the account
func (a AccountAzure) getAllScaleSets() (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	scaleSets := make(map[string]string)
	for subList, err := a.vmssClient.ListAll(ctx); subList.NotDone(); err = subList.NextWithContext(ctx) {
		if err != nil {
			return scaleSets, err
		}

		for _, scaleSet := range subList.Values() {
			rg := resourceGroupFromAzureID(*scaleSet.ID)
			if rg == "" {
				return scaleSets, fmt.Errorf("Error parsing out resource group from %s", *scaleSet.ID)
			}

			scaleSets[rg] = *scaleSet.Name
		}
	}
	return scaleSets, nil
}

// checkScaleSetVM will check all Azure Virtual Machine Scale Set instances
func (a AccountAzure) checkScaleSetVM(instanceID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Pull all the scale sets in subscription
	scaleSets, err := a.getAllScaleSets()
	if err != nil {
		return false, err
	}

	for k, v := range scaleSets {
		for subList, err := a.vmssVMClient.List(ctx, k, v, "", "", ""); subList.NotDone(); err = subList.NextWithContext(ctx) {
			if err != nil {
				return false, err
			}

			for _, instance := range subList.Values() {
				// Check for the dns_zone tag
				if val, ok := instance.Tags[a.Domain]; ok {
					if fmt.Sprintf("%s.%s", *instance.OsProfile.ComputerName, *val) == instanceID {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

// resourceGroupFromAzureID will try to parse the resource group out of a url
func resourceGroupFromAzureID(str string) string {
	match := rgRegexp.FindStringSubmatch(str)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
