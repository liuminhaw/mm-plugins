package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type userResource struct {
	client    *iam.Client
	paginator *iam.ListUsersPaginator
}

func (u *userResource) fetchConf(input any) error {
	u.paginator = iam.NewListUsersPaginator(u.client, &iam.ListUsersInput{})

	return nil
}

func (u *userResource) generate(mem *caching) (shared.MinerResource, error) {
	resource := shared.MinerResource{Identifier: iamUser}

	if err := u.fetchConf(nil); err != nil {
		return resource, fmt.Errorf("generate userResource: %w", err)
	}

	mem.usernames = []string{}
	for u.paginator.HasMorePages() {
		page, err := u.paginator.NextPage(context.Background())
		if err != nil {
			return resource, fmt.Errorf("generate userResource: failed to list iam users, %w", err)
		}

		for _, user := range page.Users {
			mem.usernames = append(mem.usernames, aws.ToString(user.UserName))
			property := shared.MinerProperty{
				Type: "User",
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(user.UserId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(user); err != nil {
				return resource, fmt.Errorf("generate userResource: %w", err)
			}

			resource.Properties = append(resource.Properties, property)
		}
	}

	return resource, nil
}
