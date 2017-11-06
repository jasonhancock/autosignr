package autosignr

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AccountAWS struct {
	Name     string   `yaml:"name"`
	Key      string   `yaml:"key"`
	Secret   string   `yaml:"secret"`
	Regions  []string `yaml:"regions"`
	awsCreds *credentials.Credentials
}

func (a *AccountAWS) Init() error {
	a.awsCreds = credentials.NewStaticCredentials(
		a.Key,
		a.Secret,
		"")

	return nil
}

func (a *AccountAWS) Type() string {
	return "aws"
}

func (a *AccountAWS) Check(instanceId string) bool {
	return a.GetInstance(instanceId) != nil
}

func (a *AccountAWS) GetInstance(instanceId string) *ec2.Instance {
	for _, region := range a.Regions {

		log.WithFields(log.Fields{
			"instance": instanceId,
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
					Name: aws.String("instance-id"),
					Values: []*string{
						aws.String(instanceId),
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

		if len(resp.Reservations) == 0 {
			log.WithFields(log.Fields{
				"instance": instanceId,
				"region":   region,
				"account":  a.Name,
				"found":    false,
			}).Debug("check-result")
			continue
		}

		for i := range resp.Reservations {
			for j := range resp.Reservations[i].Instances {
				if *resp.Reservations[i].Instances[j].InstanceId == instanceId {
					log.WithFields(log.Fields{
						"instance": instanceId,
						"region":   region,
						"account":  a.Name,
						"found":    true,
					}).Debug("check-result")
					return resp.Reservations[i].Instances[j]
				}
			}
		}
	}

	return nil
}

func (a *AccountAWS) String() string {
	return fmt.Sprintf("aws account: %s", a.Name)
}
