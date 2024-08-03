package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type lifecycleMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketLifecycleConfigurationOutput
}

func newLifecycleMiner(serviceClient utils.Client, property string) (*lifecycleMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newLifecycleMiner: %w", err)
	}

	return &lifecycleMiner{propertyType: property, serviceClient: client}, nil
}

func (l *lifecycleMiner) PropertyType() string { return l.propertyType }

func (l *lifecycleMiner) FetchConf(input any) error {
	lifecycleConfigInput, ok := input.(*s3.GetBucketLifecycleConfigurationInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketLifecycleConfigurationInput type assertion failed")
	}

	var err error
	l.configuration, err = l.serviceClient.client.GetBucketLifecycleConfiguration(
		context.Background(),
		lifecycleConfigInput,
	)
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "NoSuchLifecycleConfiguration":
				return &utils.MMError{Category: lifecycle, Code: utils.NoConfig}
			default:
				return fmt.Errorf("fetchConf lifecycle: %w", err)
			}
		}
		return fmt.Errorf("fetchConf lifecycle: %w", err)
	}

	return nil
}

func (l *lifecycleMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	err := l.FetchConf(
		&s3.GetBucketLifecycleConfigurationInput{Bucket: l.serviceClient.bucket.Name},
	)
	if err != nil {
		return nil, fmt.Errorf("generate lifecycleProp: %w", err)
	}

	for _, rule := range l.configuration.Rules {
		property := shared.MinerProperty{
			Type: lifecycle,
			Label: shared.MinerPropertyLabel{
				Name:   aws.ToString(rule.ID),
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(rule); err != nil {
			return nil, fmt.Errorf("generate lifecycleProp: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}
