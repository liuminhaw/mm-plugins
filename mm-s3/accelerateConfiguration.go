package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type accelerateMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketAccelerateConfigurationOutput
}

func newAccelerateMiner(serviceClient utils.Client, property string) (*accelerateMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newAccelerateMiner: %w", err)
	}

	return &accelerateMiner{propertyType: property, serviceClient: client}, nil
}

func (a *accelerateMiner) PropertyType() string { return a.propertyType }

func (a *accelerateMiner) FetchConf(input any) error {
	accelerateConfigInput, ok := input.(*s3.GetBucketAccelerateConfigurationInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketAccelerateConfigurationInput type assertion failed")
	}

	var err error
	a.configuration, err = a.serviceClient.client.GetBucketAccelerateConfiguration(
		context.Background(),
		accelerateConfigInput,
	)
	if err != nil {
		return fmt.Errorf("getAccelerateProperty: %w", err)
	}

	return nil
}

func (a *accelerateMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := a.FetchConf(&s3.GetBucketAccelerateConfigurationInput{Bucket: a.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate accelerate configuration: %w", err)
	}

	property := shared.MinerProperty{
		Type: accelerateConfig,
		Label: shared.MinerPropertyLabel{
			Name:   accelerateConfig,
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(a.configuration); err != nil {
		return nil, fmt.Errorf("generate accelerate configuration: %w", err)
	}
	properties = append(properties, property)

	return properties, nil
}
