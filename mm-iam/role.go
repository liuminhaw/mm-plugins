package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type roleResource struct {
	serviceClient *iamClient
}

func newRoleResource(serviceClient utils.Client) (utils.Crawler, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newRoleResource: %v", err)
	}

	return &roleResource{serviceClient: client}, nil
}

func (r *roleResource) FetchConf(input any) error {
	return nil
}

func (r *roleResource) Generate(datum utils.CacheInfo) (shared.MinerResource, error) {
	identifier := fmt.Sprintf("Role_%s", datum.Id)
	return utils.GetProperties(r.serviceClient, identifier, datum, rolePropsCrawlerConstructors)
}

// role detail (GetRole)
type roleDetailMiner struct {
	propertyType  string
	serviceClient *iamClient
	configuration *iam.GetRoleOutput
}

func newRoleDetailMiner(serviceClient utils.Client) (*roleDetailMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newRoleDetailMiner: %v", err)
	}

	return &roleDetailMiner{
		propertyType:  roleDetail,
		serviceClient: client,
	}, nil
}

func (rd *roleDetailMiner) PropertyType() string { return rd.propertyType }

func (rd *roleDetailMiner) FetchConf(input any) error {
	roleDetailInput, ok := input.(*iam.GetRoleInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetRoleInput type assertion failed")
	}

	var err error
	rd.configuration, err = rd.serviceClient.client.GetRole(context.Background(), roleDetailInput)
	if err != nil {
		return fmt.Errorf("fetchConf: %w", err)
	}

	return nil
}

func (rd *roleDetailMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := rd.FetchConf(&iam.GetRoleInput{RoleName: aws.String(datum.Name)}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate roleDetail: %w", err)
	}

	// Url decode on document content
	decodeDocument, err := utils.DocumentUrlDecode(
		aws.ToString(rd.configuration.Role.AssumeRolePolicyDocument),
	)
	if err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate roleDetail: %w", err)
	}
	rd.configuration.Role.AssumeRolePolicyDocument = aws.String(decodeDocument)

	property := shared.MinerProperty{
		Type: roleDetail,
		Label: shared.MinerPropertyLabel{
			Name:   "RoleDetail",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(rd.configuration.Role); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate roleDetail: %w", err)
	}
	properties = append(properties, property)

	return properties, nil
}

// role inline policy (GetRolePolicy)
type roleInlinePolicyMiner struct {
	propertyType  string
	serviceClient *iamClient
	paginator     *iam.ListRolePoliciesPaginator
	configuration *iam.GetRolePolicyOutput
}

func newRoleInlinePolicyMiner(serviceClient utils.Client) (*roleInlinePolicyMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newRoleInlinePolicyMiner: %v", err)
	}

	return &roleInlinePolicyMiner{
		propertyType:  roleInlinePolicy,
		serviceClient: client,
	}, nil
}

func (rip *roleInlinePolicyMiner) PropertyType() string { return rip.propertyType }

func (rip *roleInlinePolicyMiner) FetchConf(input any) error {
	roleInlinePolicyInput, ok := input.(*iam.ListRolePoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListRolePoliciesInput type assertion failed")
	}

	rip.paginator = iam.NewListRolePoliciesPaginator(
		rip.serviceClient.client,
		roleInlinePolicyInput,
	)
	return nil
}

func (rip *roleInlinePolicyMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := rip.FetchConf(&iam.ListRolePoliciesInput{RoleName: aws.String(datum.Name)}); err != nil {
		return properties, fmt.Errorf("generate roleInlinePolicy: %w", err)
	}

	for rip.paginator.HasMorePages() {
		page, err := rip.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate roleInlinePolicy: %w", err)
		}

		for _, policyName := range page.PolicyNames {
			rip.configuration, err = rip.serviceClient.client.GetRolePolicy(
				context.Background(),
				&iam.GetRolePolicyInput{
					PolicyName: aws.String(policyName),
					RoleName:   aws.String(datum.Name),
				},
			)
			if err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate roleInlinePolicy: %w", err)
			}

			// Url decode on policy document
			decodedDocument, err := utils.DocumentUrlDecode(
				aws.ToString(rip.configuration.PolicyDocument),
			)
			if err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate roleInlinePolicy: %w", err)
			}
			rip.configuration.PolicyDocument = aws.String(decodedDocument)

			property := shared.MinerProperty{
				Type: roleInlinePolicy,
				Label: shared.MinerPropertyLabel{
					Name:   policyName,
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(rip.configuration); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate roleInlinePolicy: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}

// role managed policy (ListAttachedRolePolicies)
type roleManagedPolicyMiner struct {
	propertyType  string
	serviceClient *iamClient
	paginator     *iam.ListAttachedRolePoliciesPaginator
}

func newRoleManagedPolicyMiner(serviceClient utils.Client) (*roleManagedPolicyMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newRoleManagedPolicyMiner: %v", err)
	}

	return &roleManagedPolicyMiner{
		propertyType:  roleManagedPolicy,
		serviceClient: client,
	}, nil
}

func (rmp *roleManagedPolicyMiner) PropertyType() string { return rmp.propertyType }

func (rmp *roleManagedPolicyMiner) FetchConf(input any) error {
	roleManagedPolicyInput, ok := input.(*iam.ListAttachedRolePoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAttachedRolePoliciesInput type assertion failed")
	}

	rmp.paginator = iam.NewListAttachedRolePoliciesPaginator(
		rmp.serviceClient.client,
		roleManagedPolicyInput,
	)
	return nil
}

func (rmp *roleManagedPolicyMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := rmp.FetchConf(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String(datum.Name)}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate roleManagedPolicy: %w", err)
	}

	for rmp.paginator.HasMorePages() {
		page, err := rmp.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate roleManagedPolicy: %w", err)
		}

		for _, policy := range page.AttachedPolicies {
			property := shared.MinerProperty{
				Type: roleManagedPolicy,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(policy.PolicyName),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(policy); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate roleManagedPolicy: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}

// role's instance profile
type roleInstanceProfileMiner struct {
	propertyType  string
	serviceClient *iamClient
	paginator     *iam.ListInstanceProfilesForRolePaginator
}

func newRoleInstanceProfileMiner(serivceClient utils.Client) (*roleInstanceProfileMiner, error) {
	client, err := assertIAMClient(serivceClient)
	if err != nil {
		return nil, fmt.Errorf("newRoleInstanceProfileMiner: %v", err)
	}

	return &roleInstanceProfileMiner{
		propertyType:  roleInstanceProfile,
		serviceClient: client,
	}, nil
}

func (rip *roleInstanceProfileMiner) PropertyType() string { return rip.propertyType }

func (rip *roleInstanceProfileMiner) FetchConf(input any) error {
	roleInstanceProfileInput, ok := input.(*iam.ListInstanceProfilesForRoleInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListInstanceProfilesForRoleInput type assertion failed")
	}

	rip.paginator = iam.NewListInstanceProfilesForRolePaginator(
		rip.serviceClient.client,
		roleInstanceProfileInput,
	)
	return nil
}

func (rip *roleInstanceProfileMiner) Generate(
	datum utils.CacheInfo,
) ([]shared.MinerProperty, error) {
	type instanceProfileInfo struct {
		Name string `json:"name"`
		Id   string `json:"id"`
		Arn  string `json:"arn"`
	}

	properties := []shared.MinerProperty{}

	if err := rip.FetchConf(&iam.ListInstanceProfilesForRoleInput{RoleName: aws.String(datum.Name)}); err != nil {
		return nil, fmt.Errorf("generate roleInstanceProfile: %w", err)
	}

	for rip.paginator.HasMorePages() {
		page, err := rip.paginator.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("generate roleInstanceProfile: %w", err)
		}

		for _, profile := range page.InstanceProfiles {
			property := shared.MinerProperty{
				Type: roleInstanceProfile,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(profile.InstanceProfileId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(instanceProfileInfo{
				Name: aws.ToString(profile.InstanceProfileName),
				Id:   aws.ToString(profile.InstanceProfileId),
				Arn:  aws.ToString(profile.Arn),
			}); err != nil {
				return nil, fmt.Errorf("generate roleInstanceProfile: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}
