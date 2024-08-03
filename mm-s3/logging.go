package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type loggingMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketLoggingOutput
}

func newLoggingMiner(serviceClient utils.Client, property string) (*loggingMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newLoggingMiner: %w", err)
	}

	return &loggingMiner{propertyType: property, serviceClient: client}, nil
}

func (l *loggingMiner) PropertyType() string { return l.propertyType }

func (l *loggingMiner) FetchConf(input any) error {
	loggingInput, ok := input.(*s3.GetBucketLoggingInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketLoggingInput type assertion failed")
	}

	var err error
	l.configuration, err = l.serviceClient.client.GetBucketLogging(
		context.Background(),
		loggingInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf: bucket logging: %w", err)
	}

	return nil
}

func (l *loggingMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := l.FetchConf(&s3.GetBucketLoggingInput{Bucket: l.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket logging: %w", err)
	}

	if l.configuration.LoggingEnabled != nil {
		property := shared.MinerProperty{
			Type: logging,
			Label: shared.MinerPropertyLabel{
				Name:   "Logging",
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(l.configuration.LoggingEnabled); err != nil {
			return nil, fmt.Errorf("generate bucket logging: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}
