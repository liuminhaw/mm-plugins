package utils

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/liuminhaw/mist-miner/shared"
)

type CacheInfo struct {
	Name    string
	Id      string
	Content string
}

type Client interface {
	Service() string
}

type Crawler interface {
	FetchConf(any) error
	Generate(CacheInfo) (shared.MinerResource, error)
}

type CrawlerConstructor func(ctx context.Context, client Client) (Crawler, error)

func NewCrawler(
	ctx context.Context,
	serviceClient Client,
	resourceType string,
	resourcesMapping map[string]CrawlerConstructor,
) (Crawler, error) {
	constructor, ok := resourcesMapping[resourceType]
	if !ok {
		return nil, fmt.Errorf("New crawler: unknown property type: %s", resourceType)
	}
	return constructor(ctx, serviceClient)
}

type PropsCrawler interface {
	PropertyType() string
	FetchConf(any) error
	Generate(CacheInfo) ([]shared.MinerProperty, error)
}

type PropsCrawlerConstructor func(serviceClient Client) (PropsCrawler, error)

func GetProperties(
	serviceClient Client,
	identifier string,
	datum CacheInfo,
	constructors []PropsCrawlerConstructor,
) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: identifier,
	}

	for _, constructor := range constructors {
		propsCrawler, err := constructor(serviceClient)
		if err != nil {
			return shared.MinerResource{}, fmt.Errorf("GetProperties(%s): %w", identifier, err)
		}
		propertyType := propsCrawler.PropertyType()
		log.Printf("%s property: %s\n", identifier, propertyType)

		genProps, err := propsCrawler.Generate(datum)
		if err != nil {
			var configErr *MMError
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
		return shared.MinerResource{}, &MMError{identifier, noProps}
	}

	return resource, nil
}
