package autosignr

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	gcpcompute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

var zoneRegexp = regexp.MustCompile(".*/zones/(.*)")

type AccountGCP struct {
	Name            string `yaml:"name"`
	ProjectID       string `yaml:"project_id"`
	CredentialsFile string `yaml:"credentials_file"`

	vmClient *gcpcompute.Service
}

func (a *AccountGCP) Init() error {
	if a.ProjectID == "" {
		return errors.New("GCP project_id is missing")
	}

	ctx := context.TODO()
	var err error
	a.vmClient, err = gcpcompute.NewService(ctx, option.WithCredentialsFile(a.CredentialsFile))
	return err
}

func (a AccountGCP) Type() string {
	return "gcp"
}

func (a AccountGCP) Check(instanceId string) bool {
	log.WithFields(log.Fields{
		"instance":   instanceId,
		"project_id": a.ProjectID,
	}).Debug("checking")

	zones, err := a.getActiveZones()
	if err != nil {
		return false
	}

	ctx := context.Background()
	var found bool
	for _, zone := range zones {
		req := a.vmClient.Instances.List(a.ProjectID, zone)
		_ = req.Pages(ctx, func(page *gcpcompute.InstanceList) error {
			for _, instance := range page.Items {
				// When creating instance you are able to set the hostname
				if instance.Hostname == instanceId {
					found = true
					return nil
				}
				// When creating a instancegroup you can't set the hostname
				// So look for the internal name
				internalName := fmt.Sprintf("%s.%s.c.%s.internal", instance.Name, zone, a.ProjectID)
				if internalName == instanceId {
					found = true
					return nil
				}
			}
			return nil
		})
		if found {
			log.WithFields(log.Fields{
				"instance": instanceId,
				"account":   a.Name,
				"found":     true,
			}).Debug("check-result")
			return true
		}
	}

	log.WithFields(log.Fields{
		"instance": instanceId,
		"account":  a.Name,
		"found":    false,
	}).Debug("check-result")
	return false
}

func (a AccountGCP) String() string {
	return fmt.Sprintf("gcp account: %s", a.Name)
}

func (a AccountGCP) getActiveZones() ([]string, error) {
	ctx := context.Background()

	var zones []string
	// Search all regions with instance and add the zone for that region
	req := a.vmClient.Regions.List(a.ProjectID)
	err := req.Pages(ctx, func(page *gcpcompute.RegionList) error {
		for _, region := range page.Items {
			for _, quota := range region.Quotas {
				if quota.Metric == "INSTANCES" && quota.Usage > 0 {
					for _, z := range region.Zones {
						zones = append(zones, zoneName(z))
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return []string{}, errors.Wrap(err, "pulling zones in "+a.ProjectID)
	}
	return zones, nil
}

func zoneName(str string) string {
	match := zoneRegexp.FindStringSubmatch(str)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}
