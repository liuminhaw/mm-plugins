package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type accountResource struct {
	serviceClient *iamClient
}

func newAccountResource(serviceClient utils.Client) (utils.Crawler, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newAccountResource: %v", err)
	}

	return &accountResource{serviceClient: client}, nil
}

func (a *accountResource) FetchConf(input any) error {
	return nil
}

func (a *accountResource) Generate(dummy utils.CacheInfo) (shared.MinerResource, error) {
	return utils.GetProperties(a.serviceClient, "Account", dummy, accountPropsCrawlerConstructors)
}

// Account password policy
type accountPasswordPolicyMiner struct {
	propertyType  string
	serviceClient *iamClient
	configuration *iam.GetAccountPasswordPolicyOutput
}

func newAccountPasswordPolicyMiner(
	serviceClient utils.Client,
) (*accountPasswordPolicyMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newAccountPasswordPolicyMiner: %v", err)
	}

	return &accountPasswordPolicyMiner{
		propertyType:  accountPasswordPolicy,
		serviceClient: client,
	}, nil
}

func (pp *accountPasswordPolicyMiner) PropertyType() string { return pp.propertyType }

func (pp *accountPasswordPolicyMiner) FetchConf(input any) error {
	var err error
	pp.configuration, err = pp.serviceClient.client.GetAccountPasswordPolicy(
		context.Background(),
		&iam.GetAccountPasswordPolicyInput{},
	)
	if err != nil {
		return fmt.Errorf("fetchConf passwordPolicy: %w", err)
	}

	return nil
}

func (pp *accountPasswordPolicyMiner) Generate(
	dummy utils.CacheInfo,
) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := pp.FetchConf(nil); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate password policy: %w", err)
	}

	property := shared.MinerProperty{
		Type: accountPasswordPolicy,
		Label: shared.MinerPropertyLabel{
			Name:   "PasswordPolicy",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(pp.configuration.PasswordPolicy); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate account password policy: %w", err)
	}
	properties = append(properties, property)

	return properties, nil
}

// Account summary
type accountSummaryMiner struct {
	propertyType  string
	serviceClient *iamClient
	configuration *iam.GetAccountSummaryOutput
}

func newAccountSummaryMiner(serviceClient utils.Client) (*accountSummaryMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newAccountSummaryMiner: %v", err)
	}

	return &accountSummaryMiner{
		propertyType:  accountSummary,
		serviceClient: client,
	}, nil
}

func (as *accountSummaryMiner) PropertyType() string { return as.propertyType }

func (as *accountSummaryMiner) FetchConf(input any) error {
	var err error
	as.configuration, err = as.serviceClient.client.GetAccountSummary(
		context.Background(),
		&iam.GetAccountSummaryInput{},
	)
	if err != nil {
		return fmt.Errorf("fetchConf accountSummary: %w", err)
	}

	return nil
}

func (as *accountSummaryMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := as.FetchConf(nil); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate account summary: %w", err)
	}

	property := shared.MinerProperty{
		Type: accountSummary,
		Label: shared.MinerPropertyLabel{
			Name:   "AccountSummary",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(as.configuration.SummaryMap); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate account summary: %w", err)
	}
	properties = append(properties, property)

	return properties, nil
}

// Account alias
type accountAliasMiner struct {
	propertyType  string
	serviceClient *iamClient
	paginator     *iam.ListAccountAliasesPaginator
}

func newAccountAliasMiner(serviceClient utils.Client) (*accountAliasMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newAccountAliasMiner: %v", err)
	}

	return &accountAliasMiner{
		propertyType:  accountAlias,
		serviceClient: client,
	}, nil
}

func (aa *accountAliasMiner) PropertyType() string { return aa.propertyType }

func (aa *accountAliasMiner) FetchConf(input any) error {
	accountAliasInput, ok := input.(*iam.ListAccountAliasesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAccountAliasesInput type assertion failed")
	}

	aa.paginator = iam.NewListAccountAliasesPaginator(aa.serviceClient.client, accountAliasInput)
	return nil
}

func (aa *accountAliasMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := aa.FetchConf(&iam.ListAccountAliasesInput{}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate account alias: %w", err)
	}

	for aa.paginator.HasMorePages() {
		page, err := aa.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate account alias: %w", err)
		}

		for _, alias := range page.AccountAliases {
			property := shared.MinerProperty{
				Type: accountAlias,
				Label: shared.MinerPropertyLabel{
					Name:   "AccountAlias",
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatText,
				},
			}
			if err := property.FormatContentValue(alias); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate account alias: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}
