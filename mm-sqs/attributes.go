package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type attributesMiner struct {
	propertyType  string
	serviceClient *sqsClient
	configuration *sqs.GetQueueAttributesOutput
}

func newAttributesMiner(serviceClient utils.Client, property string) (*attributesMiner, error) {
	client, err := assertSqsClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newAttributesMiner: %w", err)
	}

	return &attributesMiner{propertyType: property, serviceClient: client}, nil
}

func (a *attributesMiner) PropertyType() string { return a.propertyType }

func (a *attributesMiner) FetchConf(input any) error {
	attributesInput, ok := input.(*sqs.GetQueueAttributesInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetQueueAttributesInput type assertion failed")
	}

	var err error
	a.configuration, err = a.serviceClient.client.GetQueueAttributes(
		context.Background(),
		attributesInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf: GetQueueAttributes: %w", err)
	}

	return nil
}

func (a *attributesMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	err := a.FetchConf(
		&sqs.GetQueueAttributesInput{
			QueueUrl:       aws.String(a.serviceClient.url),
			AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("generate attributes: %w", err)
	}

	for k, v := range a.configuration.Attributes {
		property := shared.MinerProperty{
			Type: attributes,
			Label: shared.MinerPropertyLabel{
				Name:   k,
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatText,
			},
		}
		if err := property.FormatContentValue(v); err != nil {
			return nil, fmt.Errorf("generate attributes: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}
