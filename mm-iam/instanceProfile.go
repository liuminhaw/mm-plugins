package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type instanceProfileResource struct {
	serviceClient *iamClient
}

func newInstanceProfileResource(serviceClient utils.Client) (utils.Crawler, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newInstanceProfileResource: %v", err)
	}

	return &instanceProfileResource{serviceClient: client}, nil
}

func (i *instanceProfileResource) FetchConf(input any) error {
	return nil
}

func (i *instanceProfileResource) Generate(datum utils.CacheInfo) (shared.MinerResource, error) {
	Identifier := fmt.Sprintf("InstanceProfile_%s", datum.Id)
	return utils.GetProperties(
		i.serviceClient,
		Identifier,
		datum,
		instanceProfilePropsCrawlerConstructors,
	)
}

// instanceProfile detail
type instanceProfileDetailMiner struct {
	propertyType  string
	serviceClient *iamClient
	configuration *iam.GetInstanceProfileOutput
}

func newInstanceProfileDetailMiner(
	serviceClient utils.Client,
) (*instanceProfileDetailMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newInstanceProfileDetailMiner: %v", err)
	}

	return &instanceProfileDetailMiner{
		propertyType:  instanceProfileDetail,
		serviceClient: client,
	}, nil
}

func (ipd *instanceProfileDetailMiner) PropertyType() string { return ipd.propertyType }

func (ipd *instanceProfileDetailMiner) FetchConf(input any) error {
	instanceProfileInput, ok := input.(*iam.GetInstanceProfileInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetInstanceProfileInput type assertion failed")
	}

	var err error
	ipd.configuration, err = ipd.serviceClient.client.GetInstanceProfile(
		context.Background(),
		instanceProfileInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf: %w", err)
	}

	return nil
}

func (ipd *instanceProfileDetailMiner) Generate(
	datum utils.CacheInfo,
) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := ipd.FetchConf(&iam.GetInstanceProfileInput{InstanceProfileName: aws.String(datum.Name)}); err != nil {
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
