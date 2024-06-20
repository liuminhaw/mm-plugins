package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type mfaDeviceResource struct {
	client    *iam.Client
	paginator *iam.ListMFADevicesPaginator
}

func (mr *mfaDeviceResource) fetchConf(input any) error {
	mfaDeviceInput, ok := input.(*iam.ListMFADevicesInput)
	if !ok {
		return fmt.Errorf("fetchConf: type assertion failed")
	}
	mr.paginator = iam.NewListMFADevicesPaginator(mr.client, mfaDeviceInput)

	return nil
}

func (mr *mfaDeviceResource) generate(mem *caching) (shared.MinerResource, error) {
	resource := shared.MinerResource{Identifier: mfaDevice}

	if mem.usernames == nil {
		return resource, fmt.Errorf("generate mfaDeviceResource: caching usernames not found")
	}

	for _, username := range mem.usernames {
		log.Printf("Get MFA devices for user: %s", username)
		if err := mr.fetchConf(&iam.ListMFADevicesInput{UserName: aws.String(username)}); err != nil {
			return resource, fmt.Errorf("generate mfaDeviceResource: %w", err)
		}
		for mr.paginator.HasMorePages() {
			page, err := mr.paginator.NextPage(context.Background())
			if err != nil {
				return resource, fmt.Errorf(
					"generate mfaDeviceResource: failed to list mfa devices %w",
					err,
				)
			}

			for _, device := range page.MFADevices {
				property := shared.MinerProperty{
					Type: mfaDevice,
					Label: shared.MinerPropertyLabel{
						Name:   aws.ToString(device.SerialNumber),
						Unique: true,
					},
					Content: shared.MinerPropertyContent{
						Format: shared.FormatJson,
					},
				}
                if err := property.FormatContentValue(device); err != nil {
                    return resource, fmt.Errorf("generate mfaDeviceResource: %w", err)
                }

                resource.Properties = append(resource.Properties, property)
			}
		}
	}

    return resource, nil
}
