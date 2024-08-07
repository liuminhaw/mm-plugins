package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/liuminhaw/mist-miner/shared"
	iamContext "github.com/liuminhaw/mm-plugins/mm-iam/context"
	"github.com/liuminhaw/mm-plugins/utils"
)

type dataCache struct {
	resource string
	caches   []utils.CacheInfo
}

type caching struct {
	users            dataCache
	groups           dataCache
	policies         dataCache
	roles            dataCache
	virtualMFAs      dataCache
	instanceProfiles dataCache
}

func newCaching() *caching {
	return &caching{
		users:            dataCache{resource: iamUser, caches: []utils.CacheInfo{}},
		groups:           dataCache{resource: iamGroup, caches: []utils.CacheInfo{}},
		policies:         dataCache{resource: iamPolicy, caches: []utils.CacheInfo{}},
		roles:            dataCache{resource: iamRole, caches: []utils.CacheInfo{}},
		virtualMFAs:      dataCache{resource: iamVirtualMFADevice, caches: []utils.CacheInfo{}},
		instanceProfiles: dataCache{resource: iamInstanceProfile, caches: []utils.CacheInfo{}},
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
	if err := c.readVirtualMFAs(ctx, client); err != nil {
		return fmt.Errorf("caching read: %w", err)
	}
	if err := c.readInstanceProfiles(client); err != nil {
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
			c.users.caches = append(c.users.caches, utils.CacheInfo{
				Name: aws.ToString(user.UserName),
				Id:   aws.ToString(user.UserId),
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
			c.groups.caches = append(c.groups.caches, utils.CacheInfo{
				Name: aws.ToString(group.GroupName),
				Id:   aws.ToString(group.GroupId),
			})
		}
	}

	return nil
}

func (c *caching) readPolicies(ctx context.Context, client *iam.Client) error {
	equipments := iamContext.Equipments(ctx)
	listPoliciesScope := utils.GetEquipAttribute(
		equipments,
		utils.EquipmentInfo{
			TargetType: policyEquipmentType,
			TargetName: "list",
			TargetAttr: "scope",
			DefaultVal: "Local",
			AcceptVals: []string{"Local", "AWS", "All"},
		},
	)
	log.Printf("listPoliciesScope: %s\n", listPoliciesScope)

	input := iam.ListPoliciesInput{Scope: types.PolicyScopeType(listPoliciesScope)}
	paginator := iam.NewListPoliciesPaginator(client, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return fmt.Errorf("caching readPolicies: %w", err)
		}

		for _, policy := range page.Policies {
			c.policies.caches = append(c.policies.caches, utils.CacheInfo{
				Name: aws.ToString(policy.Arn),
				Id:   aws.ToString(policy.PolicyId),
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
			c.roles.caches = append(c.roles.caches, utils.CacheInfo{
				Name: aws.ToString(role.RoleName),
				Id:   aws.ToString(role.RoleId),
			})
		}
	}

	return nil
}

func (c *caching) readVirtualMFAs(ctx context.Context, client *iam.Client) error {
	listVirtualMFAAssignStatus := utils.GetEquipAttribute(
		iamContext.Equipments(ctx),
		utils.EquipmentInfo{
			TargetType: virtualMFAEquipmentType,
			TargetName: "mine",
			TargetAttr: "assignmentStatus",
			DefaultVal: "Any",
			AcceptVals: []string{"Any", "Assigned", "Unassigned"},
		},
	)
	log.Printf("listVirtualMFAAssignStatus: %s\n", listVirtualMFAAssignStatus)

	input := iam.ListVirtualMFADevicesInput{
		AssignmentStatus: types.AssignmentStatusType(listVirtualMFAAssignStatus),
	}
	paginator := iam.NewListVirtualMFADevicesPaginator(client, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return fmt.Errorf("caching readVirtualMFAs: %w", err)
		}

		for _, device := range page.VirtualMFADevices {
			marshaledDevice, err := shared.JsonMarshal(device)
			if err != nil {
				return fmt.Errorf("caching readVirtualMFAs: %w", err)
			}
			normalizedDevice, err := shared.JsonNormalize(string(marshaledDevice))
			if err != nil {
				return fmt.Errorf("caching readVirtualMFAs: %w", err)
			}

			c.virtualMFAs.caches = append(c.virtualMFAs.caches, utils.CacheInfo{
				Name:    aws.ToString(device.SerialNumber),
				Id:      aws.ToString(device.SerialNumber),
				Content: string(normalizedDevice),
			})
		}
	}

	return nil
}

func (c *caching) readInstanceProfiles(client *iam.Client) error {
	paginator := iam.NewListInstanceProfilesPaginator(client, &iam.ListInstanceProfilesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return fmt.Errorf("caching readInstanceProfiles: %w", err)
		}

		for _, profile := range page.InstanceProfiles {
			c.instanceProfiles.caches = append(c.instanceProfiles.caches, utils.CacheInfo{
				Name: aws.ToString(profile.InstanceProfileName),
				Id:   aws.ToString(profile.InstanceProfileId),
			})
		}
	}

	return nil
}
