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

type ownershipControlMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketOwnershipControlsOutput
}

func newOwnershipControlMiner(
	serviceClient utils.Client,
	property string,
) (*ownershipControlMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newOwnershipControlMiner: %w", err)
	}

	return &ownershipControlMiner{propertyType: property, serviceClient: client}, nil
}

func (oc *ownershipControlMiner) PropertyType() string { return oc.propertyType }

func (oc *ownershipControlMiner) FetchConf(input any) error {
	ownershipControlInput, ok := input.(*s3.GetBucketOwnershipControlsInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketOwnershipControlsInput type assertion failed")
	}

	var err error
	oc.configuration, err = oc.serviceClient.client.GetBucketOwnershipControls(
		context.Background(),
		ownershipControlInput,
	)
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "OwnershipControlsNotFoundError":
				return &utils.MMError{Category: ownershipControl, Code: utils.NoConfig}
			default:
				return fmt.Errorf("fetchConf bucket ownershipControl: %w", err)
			}
		}
		return fmt.Errorf("fetchConf bucket ownershipControl: %w", err)
	}

	return nil
}

func (oc *ownershipControlMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := oc.FetchConf(&s3.GetBucketOwnershipControlsInput{Bucket: oc.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket ownershipControl: %w", err)
	}

	if oc.configuration.OwnershipControls != nil {
		property := shared.MinerProperty{
			Type: ownershipControl,
			Label: shared.MinerPropertyLabel{
				Name:   "OwnershipControls",
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(oc.configuration.OwnershipControls); err != nil {
			return nil, fmt.Errorf("generate bucket ownershipControlProp: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}
