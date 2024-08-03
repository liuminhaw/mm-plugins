package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type notificationMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketNotificationConfigurationOutput
}

func newNotificationMiner(serviceClient utils.Client, property string) (*notificationMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newNotificationMiner: %w", err)
	}

	return &notificationMiner{propertyType: property, serviceClient: client}, nil
}

func (n *notificationMiner) PropertyType() string { return n.propertyType }

func (n *notificationMiner) FetchConf(input any) error {
	notificationConfigInput, ok := input.(*s3.GetBucketNotificationConfigurationInput)
	if !ok {
		return fmt.Errorf(
			"fetchConf: GetBucketNotificationConfigurationInput type assertion failed",
		)
	}

	var err error
	n.configuration, err = n.serviceClient.client.GetBucketNotificationConfiguration(
		context.Background(),
		notificationConfigInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf: bucket notification: %w", err)
	}

	return nil
}

func (n *notificationMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	err := n.FetchConf(
		&s3.GetBucketNotificationConfigurationInput{Bucket: n.serviceClient.bucket.Name},
	)
	if err != nil {
		return nil, fmt.Errorf("generate notifications configuration: %w", err)
	}

	if n.notificationIsEmpty() {
		log.Println("No notifications configuration found")
	} else {
		property := shared.MinerProperty{
			Type: notification,
			Label: shared.MinerPropertyLabel{
				Name:   "Notification",
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(n.configuration); err != nil {
			return nil, fmt.Errorf("generate notificationProp: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}

func (n *notificationMiner) notificationIsEmpty() bool {
	return n.configuration.EventBridgeConfiguration == nil &&
		len(n.configuration.LambdaFunctionConfigurations) == 0 &&
		len(n.configuration.QueueConfigurations) == 0 &&
		len(n.configuration.TopicConfigurations) == 0
}
