package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type groupResource struct {
	client *iam.Client
}

func newGroupResource(client *iam.Client) crawler {
	resource := groupResource{
		client: client,
	}
	return &resource
}

func (g *groupResource) fetchConf(input any) error {
	return nil
}

func (g *groupResource) generate(datum cacheInfo) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: fmt.Sprintf("Group_%s", datum.id),
	}

	for _, prop := range miningGroupProps {
		log.Printf("group property: %s\n", prop)

		groupPropsCrawler, err := newPropsCrawler(g.client, prop)
		if err != nil {
			return resource, fmt.Errorf("generate groupResource: %w", err)
		}
		groupProps, err := groupPropsCrawler.generate(datum)
		if err != nil {
			var configErr *mmIAMError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", prop)
			} else {
				return resource, fmt.Errorf("generate groupResource: %w", err)
			}
		} else {
			resource.Properties = append(resource.Properties, groupProps...)
		}
	}

	return resource, nil
}

// group detail
// Including information about the group and its users
type groupDetailMiner struct {
	client        *iam.Client
	configuration *iam.GetGroupOutput
}

func newGroupDetailMiner(client *iam.Client) *groupDetailMiner {
	return &groupDetailMiner{
		client: client,
	}
}

func (gd *groupDetailMiner) fetchConf(input any) error {
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

func (gd *groupDetailMiner) generate(datum cacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := gd.fetchConf(&iam.GetGroupInput{GroupName: aws.String(datum.name)}); err != nil {
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
	client        *iam.Client
	paginator     *iam.ListGroupPoliciesPaginator
	configuration *iam.GetGroupPolicyOutput
}

func newGroupInlinePolicyMiner(client *iam.Client) *groupInlinePolicyMiner {
	return &groupInlinePolicyMiner{
		client: client,
	}
}

func (gip *groupInlinePolicyMiner) fetchConf(input any) error {
	groupInlinePolicyInput, ok := input.(*iam.ListGroupPoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListGroupPoliciesInput type assertion failed")
	}

	gip.paginator = iam.NewListGroupPoliciesPaginator(gip.client, groupInlinePolicyInput)
	return nil
}

func (gip *groupInlinePolicyMiner) generate(datum cacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := gip.fetchConf(&iam.ListGroupPoliciesInput{GroupName: aws.String(datum.name)}); err != nil {
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
					GroupName:  aws.String(datum.name),
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
	client    *iam.Client
	paginator *iam.ListAttachedGroupPoliciesPaginator
}

func newGroupManagedPolicyMiner(client *iam.Client) *groupManagedPolicyMiner {
	return &groupManagedPolicyMiner{
		client: client,
	}
}

func (gmp *groupManagedPolicyMiner) fetchConf(input any) error {
	groupManagedPolicyInput, ok := input.(*iam.ListAttachedGroupPoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAttachedGroupPoliciesInput type assertion failed")
	}

	gmp.paginator = iam.NewListAttachedGroupPoliciesPaginator(gmp.client, groupManagedPolicyInput)
	return nil
}

func (gmp *groupManagedPolicyMiner) generate(datum cacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := gmp.fetchConf(&iam.ListAttachedGroupPoliciesInput{GroupName: aws.String(datum.name)}); err != nil {
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
