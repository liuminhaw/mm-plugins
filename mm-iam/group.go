package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type groupResource struct {
	client *iam.Client
}

func newGroupResource(client *iam.Client) utils.Crawler {
	resource := groupResource{
		client: client,
	}
	return &resource
}

func (g *groupResource) FetchConf(input any) error {
	return nil
}

func (g *groupResource) Generate(datum utils.CacheInfo) (shared.MinerResource, error) {
	identifier := fmt.Sprintf("Group_%s", datum.Id)
	return utils.GetProperties(g.client, identifier, datum, groupPropsCrawlerConstructors)
}

// group detail
// Including information about the group and its users
type groupDetailMiner struct {
	propertyType  string
	client        *iam.Client
	configuration *iam.GetGroupOutput
}

func newGroupDetailMiner(client *iam.Client) *groupDetailMiner {
	return &groupDetailMiner{
		propertyType: groupDetail,
		client:       client,
	}
}

func (gd *groupDetailMiner) PropertyType() string { return gd.propertyType }

func (gd *groupDetailMiner) FetchConf(input any) error {
	groupDetailInput, ok := input.(*iam.GetGroupInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetGroupInput type assertion failed")
	}

	var err error
	gd.configuration, err = gd.client.GetGroup(context.Background(), groupDetailInput)
	if err != nil {
		return fmt.Errorf("fetchConf groupDetail: %w", err)
	}

	return nil
}

func (gd *groupDetailMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := gd.FetchConf(&iam.GetGroupInput{GroupName: aws.String(datum.Name)}); err != nil {
		return properties, fmt.Errorf("generate groupDetail: %w", err)
	}

	// group detail
	property, err := gd.detailProp()
	if err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate groupDetail: %w", err)
	}
	properties = append(properties, property)

	// group users
	userProperties, err := gd.groupUserProp()
	if err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate groupDetail: %w", err)
	}
	properties = append(properties, userProperties...)

	return properties, nil
}

func (gd *groupDetailMiner) detailProp() (shared.MinerProperty, error) {
	property := shared.MinerProperty{
		Type: groupDetail,
		Label: shared.MinerPropertyLabel{
			Name:   "GroupDetail",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(gd.configuration.Group); err != nil {
		return shared.MinerProperty{}, fmt.Errorf("detailProp: %w", err)
	}

	return property, nil
}

func (gd *groupDetailMiner) groupUserProp() ([]shared.MinerProperty, error) {
	type groupUserInfo struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	}

	properties := []shared.MinerProperty{}
	for _, user := range gd.configuration.Users {
		property := shared.MinerProperty{
			Type: groupUser,
			Label: shared.MinerPropertyLabel{
				Name:   "GroupUser",
				Unique: false,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(groupUserInfo{
			Name: aws.ToString(user.UserName),
			Id:   aws.ToString(user.UserId),
		}); err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate groupUser: %w", err)
		}
		properties = append(properties, property)
	}

	return properties, nil
}

// group inline policy (ListGroupPolicies)
// Including information about the group's inline policies
type groupInlinePolicyMiner struct {
	propertyType  string
	client        *iam.Client
	paginator     *iam.ListGroupPoliciesPaginator
	configuration *iam.GetGroupPolicyOutput
}

func newGroupInlinePolicyMiner(client *iam.Client) *groupInlinePolicyMiner {
	return &groupInlinePolicyMiner{
		propertyType: groupInlinePolicy,
		client:       client,
	}
}

func (gip *groupInlinePolicyMiner) PropertyType() string { return gip.propertyType }

func (gip *groupInlinePolicyMiner) FetchConf(input any) error {
	groupInlinePolicyInput, ok := input.(*iam.ListGroupPoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListGroupPoliciesInput type assertion failed")
	}

	gip.paginator = iam.NewListGroupPoliciesPaginator(gip.client, groupInlinePolicyInput)
	return nil
}

func (gip *groupInlinePolicyMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := gip.FetchConf(&iam.ListGroupPoliciesInput{GroupName: aws.String(datum.Name)}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate groupInlinePolicy: %w", err)
	}

	for gip.paginator.HasMorePages() {
		page, err := gip.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate groupInlinePolicy: %w", err)
		}

		for _, policyName := range page.PolicyNames {
			gip.configuration, err = gip.client.GetGroupPolicy(
				context.Background(),
				&iam.GetGroupPolicyInput{
					GroupName:  aws.String(datum.Name),
					PolicyName: aws.String(policyName),
				},
			)
			if err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate groupInlinePolicy: %w", err)
			}

			// Url decode on policy document
			decodedDocument, err := utils.DocumentUrlDecode(
				aws.ToString(gip.configuration.PolicyDocument),
			)
			if err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate groupInlinePolicy: %w", err)
			}
			gip.configuration.PolicyDocument = aws.String(decodedDocument)

			property := shared.MinerProperty{
				Type: groupInlinePolicy,
				Label: shared.MinerPropertyLabel{
					Name:   policyName,
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(gip.configuration); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate groupInlinePolicy: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}

// group managed policy (ListAttachedGroupPolicies)
// Including information about the group's attached managed policies
type groupManagedPolicyMiner struct {
	propertyType string
	client       *iam.Client
	paginator    *iam.ListAttachedGroupPoliciesPaginator
}

func newGroupManagedPolicyMiner(client *iam.Client) *groupManagedPolicyMiner {
	return &groupManagedPolicyMiner{
		propertyType: groupManagedPolicy,
		client:       client,
	}
}

func (gmp *groupManagedPolicyMiner) PropertyType() string { return gmp.propertyType }

func (gmp *groupManagedPolicyMiner) FetchConf(input any) error {
	groupManagedPolicyInput, ok := input.(*iam.ListAttachedGroupPoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAttachedGroupPoliciesInput type assertion failed")
	}

	gmp.paginator = iam.NewListAttachedGroupPoliciesPaginator(gmp.client, groupManagedPolicyInput)
	return nil
}

func (gmp *groupManagedPolicyMiner) Generate(
	datum utils.CacheInfo,
) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := gmp.FetchConf(&iam.ListAttachedGroupPoliciesInput{GroupName: aws.String(datum.Name)}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate groupManagedPolicy: %w", err)
	}

	for gmp.paginator.HasMorePages() {
		page, err := gmp.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate groupManagedPolicy: %w", err)
		}

		for _, policy := range page.AttachedPolicies {
			property := shared.MinerProperty{
				Type: groupManagedPolicy,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(policy.PolicyName),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(policy); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate groupManagedPolicy: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}
