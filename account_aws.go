package autosignr

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

// AccountAWS encapsulates the account information for AWS
type AccountAWS struct {
	Name      string   `yaml:"name"`
	Key       string   `yaml:"key"`
	Secret    string   `yaml:"secret"`
	Regions   []string `yaml:"regions"`
	Attribute string   `yaml:"attribute"`
	awsCreds  *credentials.Credentials
}

// Init setup the account
func (a *AccountAWS) Init() error {
	if a.Attribute == "" {
		a.Attribute = "instance-id"
	}
	a.awsCreds = credentials.NewStaticCredentials(
		a.Key,
		a.Secret,
		"")

	return nil
}

// Type returns the type of account
func (a *AccountAWS) Type() string {
	return "aws"
}

// Check look for the instanceID in the account
func (a *AccountAWS) Check(instanceID string) bool {
	for _, region := range a.Regions {

		log.WithFields(log.Fields{
			"instance": instanceID,
			"region":   region,
			"account":  a.Name,
		}).Debug("checking")

		svc := ec2.New(session.New(), &aws.Config{
			Credentials: a.awsCreds,
			Region:      aws.String(region),
		})

		params := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String(a.Attribute),
					Values: []*string{
						aws.String(instanceID),
					},
				},
			},
		}

		// Call the DescribeInstances Operation
		resp, err := svc.DescribeInstances(params)
		if err != nil {
			log.Println(err)
			continue
		}

		found := len(resp.Reservations) > 0

		log.WithFields(log.Fields{
			"instance": instanceID,
			"region":   region,
			"account":  a.Name,
			"found":    found,
		}).Debug("check-result")

		if found {
			return true
		}
	}

	return false
}

// String returns the account info
func (a *AccountAWS) String() string {
	return fmt.Sprintf("aws account: %s", a.Name)
}
