package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type corsMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketCorsOutput
}

func newCorsMiner(serviceClient utils.Client, property string) (*corsMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newCorsMiner: %w", err)
	}

	return &corsMiner{propertyType: property, serviceClient: client}, nil
}

func (c *corsMiner) PropertyType() string { return c.propertyType }

func (c *corsMiner) FetchConf(input any) error {
	bucketCorsInput, ok := input.(*s3.GetBucketCorsInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketCorsInput type assertion failed")
	}

	var err error
	c.configuration, err = c.serviceClient.client.GetBucketCors(
		context.Background(),
		bucketCorsInput,
	)
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "NoSuchCORSConfiguration":
				return &utils.MMError{Category: cors, Code: utils.NoConfig}
			default:
				return fmt.Errorf("fetchConf bucket cors: %w", err)
			}
		}
		return fmt.Errorf("fetchConf bucket cors: %w", err)
	}

	return nil
}

func (c *corsMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := c.FetchConf(&s3.GetBucketCorsInput{Bucket: c.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket cors: %w", err)
	}
	for _, rule := range c.configuration.CORSRules {
		sortCorsRule(&rule)

		property := shared.MinerProperty{
			Type: cors,
			Label: shared.MinerPropertyLabel{
				Unique: false,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(rule); err != nil {
			return nil, fmt.Errorf("generate bucket cors: %w", err)
		}

		h := md5.New()
		h.Write([]byte(property.Content.Value))
		property.Label.Name = fmt.Sprintf("%x", h.Sum(nil))

		properties = append(properties, property)
	}

	return properties, nil
}

func sortCorsRule(rule *types.CORSRule) {
	sort.Strings(rule.AllowedMethods)
	sort.Strings(rule.AllowedOrigins)
	if rule.AllowedHeaders != nil {
		sort.Strings(rule.AllowedHeaders)
	}
	if rule.ExposeHeaders != nil {
		sort.Strings(rule.ExposeHeaders)
	}
}
