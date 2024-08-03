package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type encryptionMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketEncryptionOutput
}

func newEncryptionMiner(serviceClient utils.Client, property string) (*encryptionMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newEncryptionMiner: %w", err)
	}

	return &encryptionMiner{propertyType: property, serviceClient: client}, nil
}

func (e *encryptionMiner) PropertyType() string { return e.propertyType }

func (e *encryptionMiner) FetchConf(input any) error {
	bucketEncryptionInput, ok := input.(*s3.GetBucketEncryptionInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketEncryptionInput type assertion failed")
	}

	var err error
	e.configuration, err = e.serviceClient.client.GetBucketEncryption(
		context.Background(),
		bucketEncryptionInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf bucket encryption: %w", err)
	}

	return nil
}

func (e *encryptionMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := e.FetchConf(&s3.GetBucketEncryptionInput{Bucket: e.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket encryption: %w", err)
	}

	for _, rule := range e.configuration.ServerSideEncryptionConfiguration.Rules {
		property := shared.MinerProperty{
			Type: encryption,
			Label: shared.MinerPropertyLabel{
				Name:   "Rule",
				Unique: false,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(rule); err != nil {
			return nil, fmt.Errorf("generate bucket encryption: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}
