package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type policyResource struct {
	client *iam.Client
}

func newPolicyResource(client *iam.Client) utils.Crawler {
	resource := policyResource{
		client: client,
	}
	return &resource
}

func (p *policyResource) FetchConf(input any) error {
	return nil
}

func (p *policyResource) Generate(datum utils.CacheInfo) (shared.MinerResource, error) {
	identifier := fmt.Sprintf("Policy_%s", datum.Id)
	return utils.GetProperties(p.client, identifier, datum, policyPropsCrawlerConstructors)
}

// policy detail
type policyDetailMiner struct {
	propertyType  string
	client        *iam.Client
	configuration *iam.GetPolicyOutput
}

func newPolicyDetailMiner(client *iam.Client) *policyDetailMiner {
	return &policyDetailMiner{
		propertyType: policyDetail,
		client:       client,
	}
}

func (pd *policyDetailMiner) PropertyType() string { return pd.propertyType }

func (pd *policyDetailMiner) FetchConf(input any) error {
	policyDetailInput, ok := input.(*iam.GetPolicyInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetPolicyInput type assertion failed")
	}

	var err error
	pd.configuration, err = pd.client.GetPolicy(context.Background(), policyDetailInput)
	if err != nil {
		return fmt.Errorf("fetchConf: %w", err)
	}

	return nil
}

func (pd *policyDetailMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := pd.FetchConf(&iam.GetPolicyInput{PolicyArn: aws.String(datum.Name)}); err != nil {
		return properties, fmt.Errorf("generate policyDetail: %w", err)
	}

	property := shared.MinerProperty{
		Type: policyDetail,
		Label: shared.MinerPropertyLabel{
			Name:   "PolicyDetail",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(pd.configuration.Policy); err != nil {
		return properties, fmt.Errorf("generate policyDetail: %w", err)
	}
	properties = append(properties, property)

	return properties, nil
}

// policy versions
type policyVersionsMiner struct {
	propertyType  string
	client        *iam.Client
	configuration *iam.GetPolicyVersionOutput
	paginator     *iam.ListPolicyVersionsPaginator
}

func newPolicyVersionsMiner(client *iam.Client) *policyVersionsMiner {
	return &policyVersionsMiner{
		propertyType: policyVersions,
		client:       client,
	}
}

func (pv *policyVersionsMiner) PropertyType() string { return pv.propertyType }

func (pv *policyVersionsMiner) FetchConf(input any) error {
	policyVersionsInput, ok := input.(*iam.ListPolicyVersionsInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListPolicyVersionsInput type assertion failed")
	}

	pv.paginator = iam.NewListPolicyVersionsPaginator(pv.client, policyVersionsInput)
	return nil
}

func (pv *policyVersionsMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := pv.FetchConf(&iam.ListPolicyVersionsInput{PolicyArn: aws.String(datum.Name)}); err != nil {
		return properties, fmt.Errorf("generate policyVersions: %w", err)
	}

	for pv.paginator.HasMorePages() {
		page, err := pv.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate policyVersions: %w", err)
		}

		for _, version := range page.Versions {
			pv.configuration, err = pv.client.GetPolicyVersion(
				context.Background(),
				&iam.GetPolicyVersionInput{
					PolicyArn: aws.String(datum.Name),
					VersionId: version.VersionId,
				},
			)
			if err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate policyVersions: %w", err)
			}

			// Url decode policy document
			decodedDocument, err := utils.DocumentUrlDecode(
				aws.ToString(pv.configuration.PolicyVersion.Document),
			)
			if err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate policyVersions: %w", err)
			}
			pv.configuration.PolicyVersion.Document = aws.String(decodedDocument)

			property := shared.MinerProperty{
				Type: policyVersions,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(version.VersionId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(pv.configuration.PolicyVersion); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate policyVersions: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}
