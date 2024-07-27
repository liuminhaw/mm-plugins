package utils

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type CacheInfo struct {
	Name    string
	Id      string
	Content string
}

type Crawler interface {
	FetchConf(any) error
	Generate(CacheInfo) (shared.MinerResource, error)
}

type CrawlerConstructor func(ctx context.Context, client *iam.Client) Crawler

func NewCrawler(
	ctx context.Context,
	client *iam.Client,
	resourceType string,
	resourcesMapping map[string]CrawlerConstructor,
) (Crawler, error) {
	constructor, ok := resourcesMapping[resourceType]
	if !ok {
		return nil, fmt.Errorf("New crawler: unknown property type: %s", resourceType)
	}
	return constructor(ctx, client), nil
}

type PropsCrawler interface {
	PropertyType() string
	FetchConf(any) error
	Generate(CacheInfo) ([]shared.MinerProperty, error)
}

type PropsCrawlerConstructor func(client *iam.Client) PropsCrawler

func GetProperties(
	client *iam.Client,
	identifier string,
	datum CacheInfo,
	constructors []PropsCrawlerConstructor,
) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: identifier,
	}

	for _, constructor := range constructors {
		propsCrawler := constructor(client)
		propertyType := propsCrawler.PropertyType()
		log.Printf("%s property: %s\n", identifier, propertyType)

		genProps, err := propsCrawler.Generate(datum)
		if err != nil {
			var configErr *mmError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", propertyType)
			} else {
				return shared.MinerResource{}, fmt.Errorf("GetProperties(%s): %w", identifier, err)
			}
		} else {
			resource.Properties = append(resource.Properties, genProps...)
		}
	}

	// Check if there are any properties
	if resource.Properties == nil || len(resource.Properties) == 0 {
		return shared.MinerResource{}, &mmError{identifier, noProps}
	}

	return resource, nil
}
