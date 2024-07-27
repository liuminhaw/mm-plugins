package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type accountResource struct {
	client *iam.Client
}

func newAccountResource(client *iam.Client) utils.Crawler {
	resource := accountResource{
		client: client,
	}
	return &resource
}

func (a *accountResource) FetchConf(input any) error {
	return nil
}

func (a *accountResource) Generate(dummy utils.CacheInfo) (shared.MinerResource, error) {
	return utils.GetProperties(a.client, "Account", dummy, accountPropsCrawlerConstructors)
}

// Account password policy
type accountPasswordPolicyMiner struct {
	propertyType  string
	client        *iam.Client
	configuration *iam.GetAccountPasswordPolicyOutput
}

func newAccountPasswordPolicyMiner(client *iam.Client) *accountPasswordPolicyMiner {
	resource := accountPasswordPolicyMiner{
		propertyType: accountPasswordPolicy,
		client:       client,
	}
	return &resource
}

func (pp *accountPasswordPolicyMiner) PropertyType() string { return pp.propertyType }

func (pp *accountPasswordPolicyMiner) FetchConf(input any) error {
	var err error
	pp.configuration, err = pp.client.GetAccountPasswordPolicy(
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
	client        *iam.Client
	configuration *iam.GetAccountSummaryOutput
}

func newAccountSummaryMiner(client *iam.Client) *accountSummaryMiner {
	resource := accountSummaryMiner{
		propertyType: accountSummary,
		client:       client,
	}
	return &resource
}

func (as *accountSummaryMiner) PropertyType() string { return as.propertyType }

func (as *accountSummaryMiner) FetchConf(input any) error {
	var err error
	as.configuration, err = as.client.GetAccountSummary(
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
	propertyType string
	client       *iam.Client
	paginator    *iam.ListAccountAliasesPaginator
}

func newAccountAliasMiner(client *iam.Client) *accountAliasMiner {
	resource := accountAliasMiner{
		propertyType: accountAlias,
		client:       client,
	}
	return &resource
}

func (aa *accountAliasMiner) PropertyType() string { return aa.propertyType }

func (aa *accountAliasMiner) FetchConf(input any) error {
	accountAliasInput, ok := input.(*iam.ListAccountAliasesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAccountAliasesInput type assertion failed")
	}

	aa.paginator = iam.NewListAccountAliasesPaginator(aa.client, accountAliasInput)
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
