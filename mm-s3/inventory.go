package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type inventoryMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.ListBucketInventoryConfigurationsOutput
	requestToken  string
}

func newInventoryMiner(serviceClient utils.Client, property string) (*inventoryMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newInventoryMiner: %w", err)
	}

	return &inventoryMiner{propertyType: property, serviceClient: client}, nil
}

func (i *inventoryMiner) PropertyType() string { return i.propertyType }

func (i *inventoryMiner) FetchConf(input any) error {
	inventoryConfigInput, ok := input.(*s3.ListBucketInventoryConfigurationsInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListBucketInventoryConfigurationsInput type assertion failed")
	}

	var err error
	i.configuration, err = i.serviceClient.client.ListBucketInventoryConfigurations(
		context.Background(),
		inventoryConfigInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf: inventory configurations: %w", err)
	}

	return nil
}

func (i *inventoryMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	i.requestToken = ""
	for {
		err := i.FetchConf(
			&s3.ListBucketInventoryConfigurationsInput{
				Bucket:            i.serviceClient.bucket.Name,
				ContinuationToken: aws.String(i.requestToken),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("generate bucket inventory: %w", err)
		}

		for _, config := range i.configuration.InventoryConfigurationList {
			property := shared.MinerProperty{
				Type: inventory,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(config.Id),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(config); err != nil {
				return nil, fmt.Errorf("generate bucket inventory: %w", err)
			}

			properties = append(properties, property)
		}

		if *i.configuration.IsTruncated {
			i.requestToken = *i.configuration.NextContinuationToken
		} else {
			break
		}
	}

	return properties, nil
}
