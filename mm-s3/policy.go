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

type policyMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketPolicyOutput
}

func newPolicyMiner(serviceClient utils.Client, property string) (*policyMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newPolicyMiner: %w", err)
	}

	return &policyMiner{propertyType: property, serviceClient: client}, nil
}

func (p *policyMiner) PropertyType() string { return p.propertyType }

func (p *policyMiner) FetchConf(input any) error {
	policyInput, ok := input.(*s3.GetBucketPolicyInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketPolicyInput type assertion failed")
	}

	var err error
	p.configuration, err = p.serviceClient.client.GetBucketPolicy(context.Background(), policyInput)
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "NoSuchBucketPolicy":
				return &utils.MMError{Category: policy, Code: utils.NoConfig}
			default:
				return fmt.Errorf("fetchConf bucket policy: %w", err)
			}
		}
		return fmt.Errorf("fetchConf bucket policy: %w", err)
	}

	return nil
}

func (p *policyMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := p.FetchConf(&s3.GetBucketPolicyInput{Bucket: p.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket policy: %w", err)
	}

	if p.configuration.Policy != nil {
		normalizedPolicy, err := shared.JsonNormalize(*p.configuration.Policy)
		if err != nil {
			return nil, fmt.Errorf("generate bucket policy: %w", err)
		}

		property := shared.MinerProperty{
			Type: policy,
			Label: shared.MinerPropertyLabel{
				Name:   "Policy",
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
				Value:  string(normalizedPolicy),
			},
		}

		properties = append(properties, property)
	}

	return properties, nil
}
