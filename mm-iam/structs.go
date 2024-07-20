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

type cacheInfo struct {
	name string
	id   string
}

type dataCache struct {
	resource string
	caches   []cacheInfo
}

type caching struct {
	users    dataCache
	groups   dataCache
	policies dataCache
	roles    dataCache
}

func newCaching() *caching {
	return &caching{
		users:    dataCache{resource: iamUser, caches: []cacheInfo{}},
		groups:   dataCache{resource: iamGroup, caches: []cacheInfo{}},
		policies: dataCache{resource: iamPolicy, caches: []cacheInfo{}},
		roles:    dataCache{resource: iamRole, caches: []cacheInfo{}},
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
	if err := c.readRoles(client); err != nil {
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
			c.users.caches = append(c.users.caches, cacheInfo{
				name: aws.ToString(user.UserName),
				id:   aws.ToString(user.UserId),
			})
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
			c.groups.caches = append(c.groups.caches, cacheInfo{
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
			c.policies.caches = append(c.policies.caches, cacheInfo{
				name: aws.ToString(policy.Arn),
				id:   aws.ToString(policy.PolicyId),
			})
		}
	}

	return nil
}

func (c *caching) readRoles(client *iam.Client) error {
	paginator := iam.NewListRolesPaginator(client, &iam.ListRolesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return fmt.Errorf("caching readRoles: %w", err)
		}

		for _, role := range page.Roles {
			c.roles.caches = append(c.roles.caches, cacheInfo{
				name: aws.ToString(role.RoleName),
				id:   aws.ToString(role.RoleId),
			})
		}
	}

	return nil
}
