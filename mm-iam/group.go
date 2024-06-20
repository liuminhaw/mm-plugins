package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type groupResource struct {
	client    *iam.Client
	paginator *iam.ListGroupsPaginator
}

func (gr *groupResource) fetchConf(input any) error {
	gr.paginator = iam.NewListGroupsPaginator(gr.client, &iam.ListGroupsInput{})

	return nil
}

func (gr *groupResource) generate(memory *caching) (shared.MinerResource, error) {
	resource := shared.MinerResource{Identifier: iamGroup}

	if err := gr.fetchConf(nil); err != nil {
		return resource, fmt.Errorf("generate groupResource: %w", err)
	}
	for gr.paginator.HasMorePages() {
		page, err := gr.paginator.NextPage(context.Background())
		if err != nil {
			return resource, fmt.Errorf(
				"generate groupResource: failed to list iam groups %w",
				err,
			)
		}

		for _, group := range page.Groups {
			property := shared.MinerProperty{
				Type: "Group",
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(group.GroupId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(group); err != nil {
				return resource, fmt.Errorf("generate groupResource: %w", err)
			}

			resource.Properties = append(resource.Properties, property)
		}
	}

	return resource, nil
}

