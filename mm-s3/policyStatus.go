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

type policyStatusMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketPolicyStatusOutput
}

func newPolicyStatusMiner(serviceClient utils.Client, property string) (*policyStatusMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newPolicyStatusMiner: %w", err)
	}

	return &policyStatusMiner{propertyType: property, serviceClient: client}, nil
}

func (ps *policyStatusMiner) PropertyType() string { return ps.propertyType }

func (ps *policyStatusMiner) FetchConf(input any) error {
	policyStatusInput, ok := input.(*s3.GetBucketPolicyStatusInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketPolicyStatusInput type assertion failed")
	}

	var err error
	ps.configuration, err = ps.serviceClient.client.GetBucketPolicyStatus(
		context.Background(),
		policyStatusInput,
	)
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "NoSuchBucketPolicy":
				return &utils.MMError{Category: policyStatus, Code: utils.NoConfig}
			default:
				return fmt.Errorf("fetchConf bucket policyStatus: %w", err)
			}
		}
		return fmt.Errorf("fetchConf bucket policyStatus: %w", err)
	}

	return nil
}

func (ps *policyStatusMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := ps.FetchConf(&s3.GetBucketPolicyStatusInput{Bucket: ps.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket policyStatus: %w", err)
	}

	if ps.configuration.PolicyStatus != nil {
		property := shared.MinerProperty{
			Type: policyStatus,
			Label: shared.MinerPropertyLabel{
				Name:   "PolicyStatus",
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(ps.configuration.PolicyStatus); err != nil {
			return nil, fmt.Errorf("generate bucket policyStatus: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}
