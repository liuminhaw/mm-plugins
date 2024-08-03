package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type taggingMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketTaggingOutput
}

func newTaggingMiner(serviceClient utils.Client, property string) (*taggingMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newTaggingMiner: %w", err)
	}

	return &taggingMiner{propertyType: property, serviceClient: client}, nil
}

func (t *taggingMiner) PropertyType() string { return t.propertyType }

func (t *taggingMiner) FetchConf(input any) error {
	taggingInput, ok := input.(*s3.GetBucketTaggingInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketTaggingInput type assertion failed")
	}

	var err error
	t.configuration, err = t.serviceClient.client.GetBucketTagging(
		context.Background(),
		taggingInput,
	)
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "NoSuchTagSet":
				return &utils.MMError{Category: tagging, Code: utils.NoConfig}
			default:
				return fmt.Errorf("fetchConf bucket taggings: %w", err)
			}
		} else {
			return fmt.Errorf("fetchConf taggingProp: %w", err)
		}
	}

	return nil
}

func (t *taggingMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := t.FetchConf(&s3.GetBucketTaggingInput{Bucket: t.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket taggings: %w", err)
	}
	for _, tag := range t.configuration.TagSet {
		property := shared.MinerProperty{
			Type: tagging,
			Label: shared.MinerPropertyLabel{
				Name:   *tag.Key,
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatText,
			},
		}
		if err := property.FormatContentValue(*tag.Value); err != nil {
			return nil, fmt.Errorf("generate taggingProp: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}
