package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type versioningMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketVersioningOutput
}

func newVersioningMiner(serviceClient utils.Client, property string) (*versioningMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newVersioningMiner: %w", err)
	}

	return &versioningMiner{propertyType: property, serviceClient: client}, nil
}

func (v *versioningMiner) PropertyType() string { return v.propertyType }

func (v *versioningMiner) FetchConf(input any) error {
	versioningInput, ok := input.(*s3.GetBucketVersioningInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketVersioningInput type assertion failed")
	}

	var err error
	v.configuration, err = v.serviceClient.client.GetBucketVersioning(
		context.Background(),
		versioningInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf: bucket versioning: %w", err)
	}

	return nil
}

func (v *versioningMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := v.FetchConf(&s3.GetBucketVersioningInput{Bucket: v.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket versioning: %w", err)
	}

	property := shared.MinerProperty{
		Type: versioning,
		Label: shared.MinerPropertyLabel{
			Name:   "Versioning",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(v.configuration); err != nil {
		return nil, fmt.Errorf("generate bucket versioning: %w", err)
	}

	properties = append(properties, property)
	return properties, nil
}
