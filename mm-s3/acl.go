package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type aclMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketAclOutput
}

func newAclMiner(serviceClient utils.Client, property string) (*aclMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newAclMiner: %w", err)
	}

	return &aclMiner{propertyType: property, serviceClient: client}, nil
}

func (a *aclMiner) PropertyType() string { return a.propertyType }

func (a *aclMiner) FetchConf(input any) error {
	bucketAclInput, ok := input.(*s3.GetBucketAclInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketAclInput type assertion failed")
	}

	var err error
	a.configuration, err = a.serviceClient.client.GetBucketAcl(context.Background(), bucketAclInput)
	if err != nil {
		return fmt.Errorf("fetchConf: bucket acl: %w", err)
	}

	return nil
}

func (a *aclMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := a.FetchConf(&s3.GetBucketAclInput{Bucket: a.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate acl: %w", err)
	}

	property := shared.MinerProperty{
		Type: acl,
		Label: shared.MinerPropertyLabel{
			Name:   "Owner",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatText,
		},
	}
	if err := property.FormatContentValue(*a.configuration.Owner.ID); err != nil {
		return nil, fmt.Errorf("generate acl: %w", err)
	}
	properties = append(properties, property)

	for _, grant := range a.configuration.Grants {
		property := shared.MinerProperty{
			Type: acl,
			Label: shared.MinerPropertyLabel{
				Name:   "Grantee",
				Unique: false,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(grant); err != nil {
			return nil, fmt.Errorf("generate acl: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}
