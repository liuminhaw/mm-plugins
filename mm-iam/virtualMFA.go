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

type virtualMFADeviceResource struct {
	client *iam.Client
}

func newVirtualMFADeviceResource(client *iam.Client) crawler {
	resource := virtualMFADeviceResource{
		client: client,
	}
	return &resource
}

func (v *virtualMFADeviceResource) fetchConf(input any) error {
	return nil
}

func (v *virtualMFADeviceResource) generate(datum cacheInfo) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: fmt.Sprintf("VirtualMFA_%s", datum.id),
	}

	for _, prop := range miningVirtualMFAProps {
		log.Printf("virtualMfa property: %s\n", prop)

		virtualMfaPropsCrawler, err := newPropsCrawler(v.client, prop)
		if err != nil {
			return shared.MinerResource{}, fmt.Errorf("generate virtualMFAResource: %w", err)
		}
		virtualMFAProps, err := virtualMfaPropsCrawler.generate(datum)
		if err != nil {
			var configErr *mmIAMError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", prop)
			} else {
				return resource, fmt.Errorf("generate virtualMFAResource: %w", err)
			}
		} else {
			resource.Properties = append(resource.Properties, virtualMFAProps...)
		}

	}

	// Check if there are any properties
	if resource.Properties == nil || len(resource.Properties) == 0 {
		return shared.MinerResource{}, &mmIAMError{"VirtualMFADevice", noProps}
	}

	return resource, nil
}

// virtualMFA devices detail
type virtualMFADeviceDetailMiner struct {
	client *iam.Client
}

func newVirtualMFADeviceDetailMiner(client *iam.Client) propsCrawler {
	return &virtualMFADeviceDetailMiner{
		client: client,
	}
}

func (vmd *virtualMFADeviceDetailMiner) fetchConf(input any) error {
	return nil
}

func (vmd *virtualMFADeviceDetailMiner) generate(datum cacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	property := shared.MinerProperty{
		Type: virtualMFADeviceDetail,
		Label: shared.MinerPropertyLabel{
			Name:   "VirtualMFADetail",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
			Value:  datum.content,
		},
	}
	properties = append(properties, property)

	return properties, nil
}

// virtualMFA device tags
type virtualMFADeviceTagsMiner struct {
	client    *iam.Client
	paginator *iam.ListMFADeviceTagsPaginator
}

func newVirtualMFADeviceTagsMiner(client *iam.Client) propsCrawler {
	return &virtualMFADeviceTagsMiner{
		client: client,
	}
}

func (vmt *virtualMFADeviceTagsMiner) fetchConf(input any) error {
	mfaDeviceTagsInput, ok := input.(*iam.ListMFADeviceTagsInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListMFADeviceTagsInput type assertion failed")
	}

	vmt.paginator = iam.NewListMFADeviceTagsPaginator(vmt.client, mfaDeviceTagsInput)
	return nil
}

func (vmt *virtualMFADeviceTagsMiner) generate(datum cacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := vmt.fetchConf(&iam.ListMFADeviceTagsInput{SerialNumber: aws.String(datum.id)}); err != nil {
		return nil, fmt.Errorf("generate virtualMFADeviceTags: %w", err)
	}

	for vmt.paginator.HasMorePages() {
		page, err := vmt.paginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("generate virtualMFADeviceTags: %w", err)
		}

		for _, tag := range page.Tags {
			property := shared.MinerProperty{
				Type: virtualMFADeviceTags,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(tag.Key),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatText,
				},
			}
			if err := property.FormatContentValue(aws.ToString(tag.Value)); err != nil {
				return nil, fmt.Errorf("generate virtualMFADeviceTags: %w", err)
			}

			properties = append(properties, property)
		}
	}

	return properties, nil
}
