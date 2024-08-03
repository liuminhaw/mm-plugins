package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

// intelligentTieringProp is a crawler for fetching s3 IntelligentTiering properties
type intelligentTieringMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.ListBucketIntelligentTieringConfigurationsOutput
	requestToken  string
}

func newIntelligentTieringMiner(
	serviceClient utils.Client,
	property string,
) (*intelligentTieringMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newIntelligentTieringMiner: %w", err)
	}

	return &intelligentTieringMiner{propertyType: property, serviceClient: client}, nil
}

func (it *intelligentTieringMiner) PropertyType() string { return it.propertyType }

// fetchConf fetches the IntelligentTiering configurations for the bucket
func (it *intelligentTieringMiner) FetchConf(input any) error {
	intellTieringInput, ok := input.(*s3.ListBucketIntelligentTieringConfigurationsInput)
	if !ok {
		return fmt.Errorf(
			"fetchConf: ListBucketIntelligentTieringConfigurationsInput type assertion failed",
		)
	}

	var err error
	it.configuration, err = it.serviceClient.client.ListBucketIntelligentTieringConfigurations(
		context.Background(),
		intellTieringInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf: bucket intelligentTiering: %w", err)
	}

	return nil
}

// generate generates the IntelligentTiering properties in MinerProperty format
// to be returned to the main miner
func (it *intelligentTieringMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	it.requestToken = ""
	for {
		err := it.FetchConf(
			&s3.ListBucketIntelligentTieringConfigurationsInput{
				Bucket:            it.serviceClient.bucket.Name,
				ContinuationToken: aws.String(it.requestToken),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("generate intelligentTiering: %w", err)
		}

		for _, config := range it.configuration.IntelligentTieringConfigurationList {
			property := shared.MinerProperty{
				Type: intelligentTiering,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(config.Id),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(config); err != nil {
				return nil, fmt.Errorf("generate intelligentTiering: %w", err)
			}
			properties = append(properties, property)
		}

		if *it.configuration.IsTruncated {
			it.requestToken = *it.configuration.NextContinuationToken
		} else {
			break
		}
	}

	return properties, nil
}
