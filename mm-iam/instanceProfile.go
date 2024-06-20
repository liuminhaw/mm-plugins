package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type instanceProfileResource struct {
	client    *iam.Client
	paginator *iam.ListInstanceProfilesPaginator
}

func (ir *instanceProfileResource) fetchConf(input any) error {
	ir.paginator = iam.NewListInstanceProfilesPaginator(ir.client, &iam.ListInstanceProfilesInput{})

	return nil
}

func (ir *instanceProfileResource) generate(memory *caching) (shared.MinerResource, error) {
	resource := shared.MinerResource{Identifier: iamInstanceProfile}

	if err := ir.fetchConf(nil); err != nil {
		return resource, fmt.Errorf("generate instanceProfileResource: %w", err)
	}
	for ir.paginator.HasMorePages() {
		page, err := ir.paginator.NextPage(context.Background())
		if err != nil {
			return resource, fmt.Errorf(
				"generate instanceProfileResource: failed to list instance profiles %w",
				err,
			)
		}

		for _, profile := range page.InstanceProfiles {
			property := shared.MinerProperty{
				Type: "InstanceProfile",
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(profile.InstanceProfileId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(profile); err != nil {
				return resource, fmt.Errorf("generate instanceProfileResource: %w", err)
			}

			resource.Properties = append(resource.Properties, property)
		}
	}

	return resource, nil
}
