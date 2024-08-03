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

type websiteMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketWebsiteOutput
}

func newWebsiteMiner(serviceClient utils.Client, property string) (*websiteMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newWebsiteMiner: %w", err)
	}

	return &websiteMiner{propertyType: property, serviceClient: client}, nil
}

func (w *websiteMiner) PropertyType() string { return w.propertyType }

func (w *websiteMiner) FetchConf(input any) error {
	websiteInput, ok := input.(*s3.GetBucketWebsiteInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketWebsiteInput type assertion failed")
	}

	var err error
	w.configuration, err = w.serviceClient.client.GetBucketWebsite(
		context.Background(),
		websiteInput,
	)
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "NoSuchWebsiteConfiguration":
				return &utils.MMError{Category: website, Code: utils.NoConfig}
			default:
				return fmt.Errorf("fetchCont website: %w", err)
			}
		}
		return fmt.Errorf("fetchCont website: %w", err)
	}

	return nil
}

func (w *websiteMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := w.FetchConf(&s3.GetBucketWebsiteInput{Bucket: w.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket website: %w", err)
	}

	property := shared.MinerProperty{
		Type: website,
		Label: shared.MinerPropertyLabel{
			Name:   "Website",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(w.configuration); err != nil {
		return nil, fmt.Errorf("generate bucket website: %w", err)
	}

	properties = append(properties, property)
	return properties, nil
}
