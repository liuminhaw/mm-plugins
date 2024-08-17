package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type taggingMiner struct {
	propertyType  string
	serviceClient *sqsClient
	configuration *sqs.ListQueueTagsOutput
}

func newTaggingMiner(serviceClient utils.Client, property string) (*taggingMiner, error) {
	client, err := assertSqsClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newTaggingMiner: %w", err)
	}

	return &taggingMiner{propertyType: property, serviceClient: client}, nil
}

func (t *taggingMiner) PropertyType() string { return t.propertyType }

func (t *taggingMiner) FetchConf(input any) error {
	taggingInput, ok := input.(*sqs.ListQueueTagsInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListQueueTagsInput type assertion failed")
	}

	var err error
	t.configuration, err = t.serviceClient.client.ListQueueTags(context.Background(), taggingInput)
	if err != nil {
		return fmt.Errorf("fetchConf: ListQueueTags: %w", err)
	}

	return nil
}

func (t *taggingMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := t.FetchConf(&sqs.ListQueueTagsInput{QueueUrl: aws.String(t.serviceClient.url)}); err != nil {
		return nil, fmt.Errorf("generate tagging: %w", err)
	}

	for k, v := range t.configuration.Tags {
		property := shared.MinerProperty{
			Type:  tagging,
			Label: shared.MinerPropertyLabel{Name: k, Unique: true},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatText,
			},
		}
		if err := property.FormatContentValue(v); err != nil {
			return nil, fmt.Errorf("generate tagging: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}
