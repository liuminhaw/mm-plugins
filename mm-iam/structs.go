package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type caching struct {
	usernames []string
}

func newCaching() *caching {
    return &caching{ 
        usernames: []string{},
    }
}

func (c *caching) readUsernames(client *iam.Client) error {
	paginator := iam.NewListUsersPaginator(client, &iam.ListUsersInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return fmt.Errorf("caching readUsernames: %w", err)
		}

		for _, user := range page.Users {
			c.usernames = append(c.usernames, aws.ToString(user.UserName))
		}
	}

    return nil
}
