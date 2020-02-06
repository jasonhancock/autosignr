package autosignr

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	log "github.com/sirupsen/logrus"
)

type AccountAzure struct {
	Name    	   string   `yaml:"name"`
	ClientID       string   `yaml:"client_id"`
	ClientSecret   string   `yaml:"client_secret"`
	SubscriptionID string   `yaml:"subscription_id"`
	TenantID       string   `yaml:"tenant_id"`
	Attribute      string   `yaml:"attribute"`

	vmClient compute.VirtualMachinesClient
}

func (a *AccountAzure) Init() error {
	if a.Attribute == "" {
		a.Attribute = "Name"
	}

	conf := auth.NewClientCredentialsConfig(a.ClientID, a.ClientSecret, a.TenantID)

	a.vmClient = compute.NewVirtualMachinesClient(a.SubscriptionID)
	var err error
	a.vmClient.Authorizer, err = conf.Authorizer()
	return err
}

func (a AccountAzure) Type() string {
	return "azure"
}

func (a *AccountAzure) Check(instanceId string) bool {
	log.WithFields(log.Fields{
		"instance": instanceId,
		"subscription": a.Name,
	}).Debug("checking")

	ctx := context.TODO()
	list, err := a.vmClient.ListAll(ctx, "")
	if err != nil {
		log.Println(err)
		return false
	}

	for list.NotDone() {
		instances := list.Values()
		for _, instance := range instances {
			for k, v := range instance.Tags {
				if k == a.Attribute && *v == instanceId {
					log.WithFields(log.Fields{
						"instance": instanceId,
						"account":  a.Name,
						"found":    true,
					}).Debug("check-result")
					return true
				}
			}
		}

		if err := list.NextWithContext(ctx); err != nil {
			log.Println(err)
			return false
		}
	}

	log.WithFields(log.Fields{
		"instance": instanceId,
		"account":  a.Name,
		"found":    false,
	}).Debug("check-result")
	return false
}

func (a AccountAzure) String() string {
	return fmt.Sprintf("azure account: %s", a.Name)
}
