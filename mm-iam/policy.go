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
	serviceClient *iamClient
}

func newPolicyResource(serviceClient utils.Client) (utils.Crawler, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newPolicyResource: %v", err)
	}

	return &policyResource{serviceClient: client}, nil
}

func (p *policyResource) FetchConf(input any) error {
	return nil
}

func (p *policyResource) Generate(datum utils.CacheInfo) (shared.MinerResource, error) {
	identifier := fmt.Sprintf("Policy_%s", datum.Id)
	return utils.GetProperties(p.serviceClient, identifier, datum, policyPropsCrawlerConstructors)
}

// policy detail
type policyDetailMiner struct {
	propertyType  string
	serviceClient *iamClient
	configuration *iam.GetPolicyOutput
}

func newPolicyDetailMiner(serviceClient utils.Client) (*policyDetailMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newPolicyDetailMiner: %v", err)
	}

	return &policyDetailMiner{
		propertyType:  policyDetail,
		serviceClient: client,
	}, nil
}

func (pd *policyDetailMiner) PropertyType() string { return pd.propertyType }

func (pd *policyDetailMiner) FetchConf(input any) error {
	policyDetailInput, ok := input.(*iam.GetPolicyInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetPolicyInput type assertion failed")
	}

	var err error
	pd.configuration, err = pd.serviceClient.client.GetPolicy(
		context.Background(),
		policyDetailInput,
	)
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
	serviceClient *iamClient
	configuration *iam.GetPolicyVersionOutput
	paginator     *iam.ListPolicyVersionsPaginator
}

func newPolicyVersionsMiner(serviceClient utils.Client) (*policyVersionsMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newPolicyVersionsMiner: %v", err)
	}

	return &policyVersionsMiner{
		propertyType:  policyVersions,
		serviceClient: client,
	}, nil
}

func (pv *policyVersionsMiner) PropertyType() string { return pv.propertyType }

func (pv *policyVersionsMiner) FetchConf(input any) error {
	policyVersionsInput, ok := input.(*iam.ListPolicyVersionsInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListPolicyVersionsInput type assertion failed")
	}

	pv.paginator = iam.NewListPolicyVersionsPaginator(pv.serviceClient.client, policyVersionsInput)
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
			pv.configuration, err = pv.serviceClient.client.GetPolicyVersion(
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
