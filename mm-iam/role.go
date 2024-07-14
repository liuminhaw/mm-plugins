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

type roleResource struct {
	client *iam.Client
}

func newRoleResource(client *iam.Client) crawler {
	resource := roleResource{
		client: client,
	}
	return &resource
}

func (r *roleResource) fetchConf(input any) error {
	return nil
}

func (r *roleResource) generate(datum cacheInfo) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: fmt.Sprintf("Role_%s", datum.id),
	}

	for _, prop := range miningRoleProps {
		log.Printf("role property: %s\n", prop)

		rolePropsCrawler, err := newPropsCrawler(r.client, prop)
		if err != nil {
			return resource, fmt.Errorf("generate roleResource: %w", err)
		}
		roleProps, err := rolePropsCrawler.generate(datum.name)
		if err != nil {
			var configErr *mmIAMError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", prop)
			} else {
				return resource, fmt.Errorf("generate roleResource: %w", err)
			}
		} else {
			resource.Properties = append(resource.Properties, roleProps...)
		}
	}

	return resource, nil
}

// role detail (GetRole)
type roleDetailMiner struct {
	client        *iam.Client
	configuration *iam.GetRoleOutput
}

func (rd *roleDetailMiner) fetchConf(input any) error {
	roleDetailInput, ok := input.(*iam.GetRoleInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetRoleInput type assertion failed")
	}

	var err error
	rd.configuration, err = rd.client.GetRole(context.Background(), roleDetailInput)
	if err != nil {
		return fmt.Errorf("fetchConf: %w", err)
	}

	return nil
}

func (rd *roleDetailMiner) generate(roleName string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := rd.fetchConf(&iam.GetRoleInput{RoleName: aws.String(roleName)}); err != nil {
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
	client        *iam.Client
	paginator     *iam.ListRolePoliciesPaginator
	configuration *iam.GetRolePolicyOutput
}

func (rip *roleInlinePolicyMiner) fetchConf(input any) error {
	roleInlinePolicyInput, ok := input.(*iam.ListRolePoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListRolePoliciesInput type assertion failed")
	}

	rip.paginator = iam.NewListRolePoliciesPaginator(rip.client, roleInlinePolicyInput)
	return nil
}

func (rip *roleInlinePolicyMiner) generate(roleName string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := rip.fetchConf(&iam.ListRolePoliciesInput{RoleName: aws.String(roleName)}); err != nil {
		return properties, fmt.Errorf("generate roleInlinePolicy: %w", err)
	}

	for rip.paginator.HasMorePages() {
		page, err := rip.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate roleInlinePolicy: %w", err)
		}

		for _, policyName := range page.PolicyNames {
			rip.configuration, err = rip.client.GetRolePolicy(
				context.Background(),
				&iam.GetRolePolicyInput{
					PolicyName: aws.String(policyName),
					RoleName:   aws.String(roleName),
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
	client    *iam.Client
	paginator *iam.ListAttachedRolePoliciesPaginator
}

func (rmp *roleManagedPolicyMiner) fetchConf(input any) error {
	roleManagedPolicyInput, ok := input.(*iam.ListAttachedRolePoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAttachedRolePoliciesInput type assertion failed")
	}

	rmp.paginator = iam.NewListAttachedRolePoliciesPaginator(rmp.client, roleManagedPolicyInput)
	return nil
}

func (rmp *roleManagedPolicyMiner) generate(roleName string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := rmp.fetchConf(&iam.ListAttachedRolePoliciesInput{RoleName: aws.String(roleName)}); err != nil {
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
