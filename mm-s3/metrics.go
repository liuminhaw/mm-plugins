package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type metricsMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.ListBucketMetricsConfigurationsOutput
	requestToken  string
}

func newMetricsMiner(serviceClient utils.Client, property string) (*metricsMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newMetricsMiner: %w", err)
	}

	return &metricsMiner{propertyType: property, serviceClient: client}, nil
}

func (m *metricsMiner) PropertyType() string { return m.propertyType }

func (m *metricsMiner) FetchConf(input any) error {
	metricsConfigInput, ok := input.(*s3.ListBucketMetricsConfigurationsInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListBucketMetricsConfigurationsInput type assertion failed")
	}

	var err error
	m.configuration, err = m.serviceClient.client.ListBucketMetricsConfigurations(
		context.Background(),
		metricsConfigInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf: metrics configurations: %w", err)
	}

	return nil
}

func (m *metricsMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	m.requestToken = ""
	for {
		err := m.FetchConf(
			&s3.ListBucketMetricsConfigurationsInput{
				Bucket:            m.serviceClient.bucket.Name,
				ContinuationToken: aws.String(m.requestToken),
			},
		)
		if err != nil {
			return nil, fmt.Errorf("generate metrics: %w", err)
		}

		for _, config := range m.configuration.MetricsConfigurationList {
			property := shared.MinerProperty{
				Type: metrics,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(config.Id),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(config); err != nil {
				return nil, fmt.Errorf("generate metrics: %w", err)
			}

			properties = append(properties, property)
		}

		if *m.configuration.IsTruncated {
			m.requestToken = *m.configuration.NextContinuationToken
		} else {
			break
		}
	}

	return properties, nil
}
