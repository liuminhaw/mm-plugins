package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type analyticsMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.ListBucketAnalyticsConfigurationsOutput
	requestToken  string
}

func newAnalyticsMiner(serviceClient utils.Client, property string) (*analyticsMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newAnalyticsMiner: %w", err)
	}

	return &analyticsMiner{propertyType: property, serviceClient: client}, nil
}

func (a *analyticsMiner) PropertyType() string { return a.propertyType }

func (a *analyticsMiner) FetchConf(input any) error {
	analyicsConfigInput, ok := input.(*s3.ListBucketAnalyticsConfigurationsInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListBucketAnalyticsConfigurationsInput type assertion failed")
	}

	var err error
	a.configuration, err = a.serviceClient.client.ListBucketAnalyticsConfigurations(
		context.Background(),
		analyicsConfigInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf: analytics configurations: %w", err)
	}

	return nil
}

func (a *analyticsMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	a.requestToken = ""
	for {
		err := a.FetchConf(
			&s3.ListBucketAnalyticsConfigurationsInput{
				Bucket:            a.serviceClient.bucket.Name,
				ContinuationToken: aws.String(a.requestToken),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("generate analytics: %w", err)
		}

		for _, config := range a.configuration.AnalyticsConfigurationList {
			property := shared.MinerProperty{
				Type: analyticsConfig,
				Label: shared.MinerPropertyLabel{
					Name:   *config.Id,
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(config); err != nil {
				return nil, fmt.Errorf("generate analytics: %w", err)
			}

			properties = append(properties, property)
		}

		if *a.configuration.IsTruncated {
			a.requestToken = *a.configuration.NextContinuationToken
		} else {
			break
		}
	}

	return properties, nil
}
