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
	Name        string
	Key         string
	Secret      string
	Regions     []string
	AccountType string
	aws_creds   *credentials.Credentials
}

func NewAccountAWS(data map[interface{}]interface{}) *AccountAWS {

	r := make([]string, len(data["regions"].([]interface{})))
	for i := range data["regions"].([]interface{}) {
		r[i] = data["regions"].([]interface{})[i].(string)
	}

	f := AccountAWS{
		Name:        data["name"].(string),
		Key:         data["key_id"].(string),
		Secret:      data["secret_key"].(string),
		Regions:     r,
		AccountType: "aws",
	}

	f.aws_creds = credentials.NewStaticCredentials(
		f.Key,
		f.Secret,
		"")

	return &f
}

func (a AccountAWS) Type() string {
	return a.AccountType
}

func (a AccountAWS) Check(instanceId string) bool {
	for _, region := range a.Regions {

		log.WithFields(log.Fields{
			"instance": instanceId,
			"region":   region,
			"account":  a.Name,
		}).Debug("checking")

		svc := ec2.New(session.New(), &aws.Config{
			Credentials: a.aws_creds,
			Region:      aws.String(region),
		})

		params := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				&ec2.Filter{
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
			panic(err)
		}

		found := len(resp.Reservations) > 0
		log.WithFields(log.Fields{
			"instance": instanceId,
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

func (a AccountAWS) String() string {
	return fmt.Sprintf("aws account: %s", a.Name)
}
