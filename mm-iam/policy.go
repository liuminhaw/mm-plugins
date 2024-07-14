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

type policyResource struct {
	client *iam.Client
}

func newPolicyResource(client *iam.Client) crawler {
	resource := policyResource{
		client: client,
	}
	return &resource
}

func (p *policyResource) fetchConf(input any) error {
	return nil
}

func (p *policyResource) generate(datum cacheInfo) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: fmt.Sprintf("Policy_%s", datum.id),
	}

	for _, prop := range miningPolicyProps {
		log.Printf("policy property: %s\n", prop)

		policyPropsCrawler, err := newPropsCrawler(p.client, prop)
		if err != nil {
			return resource, fmt.Errorf("generate policyResource: %w", err)
		}
		policyProps, err := policyPropsCrawler.generate(datum.name)
		if err != nil {
			var configErr *mmIAMError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", prop)
			} else {
				return resource, fmt.Errorf("generate policyResource: %w", err)
			}
		} else {
			resource.Properties = append(resource.Properties, policyProps...)
		}
	}

	return resource, nil
}

// policy detail
type policyDetailMiner struct {
	client        *iam.Client
	configuration *iam.GetPolicyOutput
}

func newPolicyDetailMiner(client *iam.Client) propsCrawler {
	return &policyDetailMiner{
		client: client,
	}
}

func (pd *policyDetailMiner) fetchConf(input any) error {
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

func (pd *policyDetailMiner) generate(policyArn string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := pd.fetchConf(&iam.GetPolicyInput{PolicyArn: aws.String(policyArn)}); err != nil {
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
	client        *iam.Client
	configuration *iam.GetPolicyVersionOutput
	paginator     *iam.ListPolicyVersionsPaginator
}

func newPolicyVersionsMiner(client *iam.Client) propsCrawler {
	return &policyVersionsMiner{
		client: client,
	}
}

func (pv *policyVersionsMiner) fetchConf(input any) error {
	policyVersionsInput, ok := input.(*iam.ListPolicyVersionsInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListPolicyVersionsInput type assertion failed")
	}

	pv.paginator = iam.NewListPolicyVersionsPaginator(pv.client, policyVersionsInput)
	return nil
}

func (pv *policyVersionsMiner) generate(policyArn string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := pv.fetchConf(&iam.ListPolicyVersionsInput{PolicyArn: aws.String(policyArn)}); err != nil {
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
					PolicyArn: aws.String(policyArn),
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
