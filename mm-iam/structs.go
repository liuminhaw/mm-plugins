package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamContext "github.com/liuminhaw/mm-plugins/mm-iam/context"
	"github.com/liuminhaw/mm-plugins/mm-iam/utils"
)

type dataCache struct {
	name string
	id   string
}
type (
	userCache   dataCache
	groupCache  dataCache
	policyCache struct {
		arn string
		id  string
	}
)

type caching struct {
	users    []userCache
	groups   []groupCache
	policies []policyCache
}

func newCaching() *caching {
	return &caching{
		users:    []userCache{},
		groups:   []groupCache{},
		policies: []policyCache{},
	}
}

func (c *caching) read(ctx context.Context, client *iam.Client) error {
	if err := c.readUsers(client); err != nil {
		return fmt.Errorf("caching read: %w", err)
	}
	if err := c.readGroups(client); err != nil {
		return fmt.Errorf("caching read: %w", err)
	}
	if err := c.readPolicies(ctx, client); err != nil {
		return fmt.Errorf("caching read: %w", err)
	}

	return nil
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

func (c *caching) readGroups(client *iam.Client) error {
	paginator := iam.NewListGroupsPaginator(client, &iam.ListGroupsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return fmt.Errorf("caching readGroups: %w", err)
		}

		for _, group := range page.Groups {
			c.groups = append(c.groups, groupCache{
				name: aws.ToString(group.GroupName),
				id:   aws.ToString(group.GroupId),
			})
		}
	}

	return nil
}

func (c *caching) readPolicies(ctx context.Context, client *iam.Client) error {
	equipments := iamContext.Equipments(ctx)
	listPoliciesScope, _ := utils.GetEquipAttribute(
		equipments,
		policyEquipmentType,
		"list",
		"scope",
	)

	var input iam.ListPoliciesInput
	switch listPoliciesScope {
	case "Local", "AWS", "All":
		input = iam.ListPoliciesInput{Scope: types.PolicyScopeType(listPoliciesScope)}
	default:
		input = iam.ListPoliciesInput{Scope: types.PolicyScopeType("Local")}
	}

	paginator := iam.NewListPoliciesPaginator(client, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return fmt.Errorf("caching readPolicies: %w", err)
		}

		for _, policy := range page.Policies {
			c.policies = append(c.policies, policyCache{
				arn: aws.ToString(policy.Arn),
				id:  aws.ToString(policy.PolicyId),
			})
		}
	}

	return nil
}
