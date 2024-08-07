package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type virtualMFADeviceResource struct {
	serviceClient *iamClient
}

func newVirtualMFADeviceResource(serviceClient utils.Client) (utils.Crawler, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newVirtualMFADeviceResource: %v", err)
	}

	return &virtualMFADeviceResource{serviceClient: client}, nil
}

func (v *virtualMFADeviceResource) FetchConf(input any) error {
	return nil
}

func (v *virtualMFADeviceResource) Generate(datum utils.CacheInfo) (shared.MinerResource, error) {
	Identifier := fmt.Sprintf("VirtualMFA_%s", datum.Id)
	return utils.GetProperties(
		v.serviceClient,
		Identifier,
		datum,
		virtualMFADevicePropsCrawlerConstructors,
	)
}

// virtualMFA devices detail
type virtualMFADeviceDetailMiner struct {
	propertyType  string
	serviceClient *iamClient
}

func newVirtualMFADeviceDetailMiner(
	serviceClient utils.Client,
) (*virtualMFADeviceDetailMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newVirtualMFADeviceDetailMiner: %v", err)
	}

	return &virtualMFADeviceDetailMiner{
		propertyType:  virtualMFADeviceDetail,
		serviceClient: client,
	}, nil
}

func (vmd *virtualMFADeviceDetailMiner) PropertyType() string { return vmd.propertyType }

func (vmd *virtualMFADeviceDetailMiner) FetchConf(input any) error {
	return nil
}

func (vmd *virtualMFADeviceDetailMiner) Generate(
	datum utils.CacheInfo,
) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	property := shared.MinerProperty{
		Type: virtualMFADeviceDetail,
		Label: shared.MinerPropertyLabel{
			Name:   "VirtualMFADetail",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
			Value:  datum.Content,
		},
	}
	properties = append(properties, property)

	return properties, nil
}

// virtualMFA device tags
type virtualMFADeviceTagsMiner struct {
	propertyType  string
	serviceClient *iamClient
	paginator     *iam.ListMFADeviceTagsPaginator
}

func newVirtualMFADeviceTagsMiner(serviceClient utils.Client) (*virtualMFADeviceTagsMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newVirtualMFADeviceTagsMiner: %v", err)
	}

	return &virtualMFADeviceTagsMiner{
		propertyType:  virtualMFADeviceTags,
		serviceClient: client,
	}, nil
}

func (vmt *virtualMFADeviceTagsMiner) PropertyType() string { return vmt.propertyType }

func (vmt *virtualMFADeviceTagsMiner) FetchConf(input any) error {
	mfaDeviceTagsInput, ok := input.(*iam.ListMFADeviceTagsInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListMFADeviceTagsInput type assertion failed")
	}

	vmt.paginator = iam.NewListMFADeviceTagsPaginator(vmt.serviceClient.client, mfaDeviceTagsInput)
	return nil
}

func (vmt *virtualMFADeviceTagsMiner) Generate(
	datum utils.CacheInfo,
) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := vmt.FetchConf(&iam.ListMFADeviceTagsInput{SerialNumber: aws.String(datum.Id)}); err != nil {
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
