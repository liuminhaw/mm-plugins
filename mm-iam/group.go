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
	serviceClient *iamClient
}

func newGroupResource(serviceClient utils.Client) (utils.Crawler, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newGroupResource: %v", err)
	}

	return &groupResource{serviceClient: client}, nil
}

func (g *groupResource) FetchConf(input any) error {
	return nil
}

func (g *groupResource) Generate(datum utils.CacheInfo) (shared.MinerResource, error) {
	identifier := fmt.Sprintf("Group_%s", datum.Id)
	return utils.GetProperties(g.serviceClient, identifier, datum, groupPropsCrawlerConstructors)
}

// group detail
// Including information about the group and its users
type groupDetailMiner struct {
	propertyType  string
	serviceClient *iamClient
	configuration *iam.GetGroupOutput
}

func newGroupDetailMiner(serviceClient utils.Client) (*groupDetailMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newGroupDetailMiner: %v", err)
	}

	return &groupDetailMiner{
		propertyType:  groupDetail,
		serviceClient: client,
	}, nil
}

func (gd *groupDetailMiner) PropertyType() string { return gd.propertyType }

func (gd *groupDetailMiner) FetchConf(input any) error {
	groupDetailInput, ok := input.(*iam.GetGroupInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetGroupInput type assertion failed")
	}

	var err error
	gd.configuration, err = gd.serviceClient.client.GetGroup(context.Background(), groupDetailInput)
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
	serviceClient *iamClient
	paginator     *iam.ListGroupPoliciesPaginator
	configuration *iam.GetGroupPolicyOutput
}

func newGroupInlinePolicyMiner(serviceClient utils.Client) (*groupInlinePolicyMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newGroupInlinePolicyMiner: %v", err)
	}

	return &groupInlinePolicyMiner{
		propertyType:  groupInlinePolicy,
		serviceClient: client,
	}, nil
}

func (gip *groupInlinePolicyMiner) PropertyType() string { return gip.propertyType }

func (gip *groupInlinePolicyMiner) FetchConf(input any) error {
	groupInlinePolicyInput, ok := input.(*iam.ListGroupPoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListGroupPoliciesInput type assertion failed")
	}

	gip.paginator = iam.NewListGroupPoliciesPaginator(
		gip.serviceClient.client,
		groupInlinePolicyInput,
	)
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
			gip.configuration, err = gip.serviceClient.client.GetGroupPolicy(
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
	propertyType  string
	serviceClient *iamClient
	paginator     *iam.ListAttachedGroupPoliciesPaginator
}

func newGroupManagedPolicyMiner(serviceClient utils.Client) (*groupManagedPolicyMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newGroupManagedPolicyMiner: %v", err)
	}

	return &groupManagedPolicyMiner{
		propertyType:  groupManagedPolicy,
		serviceClient: client,
	}, nil
}

func (gmp *groupManagedPolicyMiner) PropertyType() string { return gmp.propertyType }

func (gmp *groupManagedPolicyMiner) FetchConf(input any) error {
	groupManagedPolicyInput, ok := input.(*iam.ListAttachedGroupPoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAttachedGroupPoliciesInput type assertion failed")
	}

	gmp.paginator = iam.NewListAttachedGroupPoliciesPaginator(
		gmp.serviceClient.client,
		groupManagedPolicyInput,
	)
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
