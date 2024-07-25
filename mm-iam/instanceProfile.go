package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type instanceProfileResource struct {
	client *iam.Client
}

func newInstanceProfileResource(client *iam.Client) crawler {
	resource := instanceProfileResource{
		client: client,
	}
	return &resource
}

func (i *instanceProfileResource) fetchConf(input any) error {
	return nil
}

func (i *instanceProfileResource) generate(datum cacheInfo) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: fmt.Sprintf("InstanceProfile_%s", datum.id),
	}

	for _, prop := range miningInstanceProfileProps {
		log.Printf("instanceProfile property: %s\n", prop)

		instanceProfilePropsCrawler, err := newPropsCrawler(i.client, prop)
		if err != nil {
			return shared.MinerResource{}, fmt.Errorf("generate instanceProfileResource: %w", err)
		}
		instanceProfileProps, err := instanceProfilePropsCrawler.generate(datum)
		if err != nil {
			var configErr *mmIAMError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", prop)
			} else {
				return resource, fmt.Errorf("generate roleResource: %w", err)
			}
		} else {
			resource.Properties = append(resource.Properties, instanceProfileProps...)
		}
	}

	// Check if there are any properties
	if resource.Properties == nil || len(resource.Properties) == 0 {
		return shared.MinerResource{}, &mmIAMError{"InstanceProfile", noProps}
	}

	return resource, nil
}

// instanceProfile detail
type instanceProfileDetailMiner struct {
	client        *iam.Client
	configuration *iam.GetInstanceProfileOutput
}

func newInstanceProfileDetailMiner(client *iam.Client) propsCrawler {
	return &instanceProfileDetailMiner{
		client: client,
	}
}

func (ipd *instanceProfileDetailMiner) fetchConf(input any) error {
	instanceProfileInput, ok := input.(*iam.GetInstanceProfileInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetInstanceProfileInput type assertion failed")
	}

	var err error
	ipd.configuration, err = ipd.client.GetInstanceProfile(
		context.Background(),
		instanceProfileInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf: %w", err)
	}

	return nil
}

func (ipd *instanceProfileDetailMiner) generate(datum cacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := ipd.fetchConf(&iam.GetInstanceProfileInput{InstanceProfileName: aws.String(datum.name)}); err != nil {
		return nil, fmt.Errorf("generate instanceProfileDetail: %w", err)
	}

	property := shared.MinerProperty{
		Type: instanceProfileDetail,
		Label: shared.MinerPropertyLabel{
			Name:   "InstanceProfileDetail",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(ipd.configuration.InstanceProfile); err != nil {
		return nil, fmt.Errorf("generate instanceProfileDetail: %w", err)
	}
	properties = append(properties, property)

	return properties, nil
}
