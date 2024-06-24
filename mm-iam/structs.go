package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
)

type userCache struct {
	name string
	id   string
}

type caching struct {
	users []userCache
}

func newCaching() *caching {
	return &caching{
		users: []userCache{},
	}
}

func (c *caching) readUsers(client *iam.Client) error {
	paginator := iam.NewListUsersPaginator(client, &iam.ListUsersInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return fmt.Errorf("caching readUsernames: %w", err)
		}

		for _, user := range page.Users {
			c.users = append(c.users, userCache{
				name: aws.ToString(user.UserName),
				id:   aws.ToString(user.UserId),
			})
			// c.usernames = append(c.usernames, aws.ToString(user.UserName))
		}
	}

	return nil
}
