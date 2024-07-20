package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type accountResource struct {
	client *iam.Client
}

func newAccountResource(client *iam.Client) crawler {
	resource := accountResource{
		client: client,
	}
	return &resource
}

func (a *accountResource) fetchConf(input any) error {
	return nil
}

func (a *accountResource) generate(datum cacheInfo) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: "Account",
	}

	for _, prop := range miningAccountProps {
		log.Printf("account property: %s\n", prop)

		accountPropsCrawler, err := newPropsCrawler(a.client, prop)
		if err != nil {
			return shared.MinerResource{}, fmt.Errorf("generate accountResource: %w", err)
		}
		accountProps, err := accountPropsCrawler.generate("")
		if err != nil {
			var configErr *mmIAMError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", prop)
			} else {
				return shared.MinerResource{}, fmt.Errorf("generate accountResource: %w", err)
			}
		} else {
			resource.Properties = append(resource.Properties, accountProps...)
		}
	}

	return resource, nil
}

// Account password policy
type accountPasswordPolicyMiner struct {
	client        *iam.Client
	configuration *iam.GetAccountPasswordPolicyOutput
}

func newAccountPasswordPolicyMiner(client *iam.Client) *accountPasswordPolicyMiner {
	resource := accountPasswordPolicyMiner{
		client: client,
	}
	return &resource
}

func (pp *accountPasswordPolicyMiner) fetchConf(input any) error {
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

func (pp *accountPasswordPolicyMiner) generate(dummy string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := pp.fetchConf(nil); err != nil {
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
	client        *iam.Client
	configuration *iam.GetAccountSummaryOutput
}

func newAccountSummaryMiner(client *iam.Client) *accountSummaryMiner {
	resource := accountSummaryMiner{
		client: client,
	}
	return &resource
}

func (as *accountSummaryMiner) fetchConf(input any) error {
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

func (as *accountSummaryMiner) generate(dummy string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := as.fetchConf(nil); err != nil {
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
	client    *iam.Client
	paginator *iam.ListAccountAliasesPaginator
}

func newAccountAliasMiner(client *iam.Client) *accountAliasMiner {
	resource := accountAliasMiner{
		client: client,
	}
	return &resource
}

func (aa *accountAliasMiner) fetchConf(input any) error {
	accountAliasInput, ok := input.(*iam.ListAccountAliasesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAccountAliasesInput type assertion failed")
	}

	aa.paginator = iam.NewListAccountAliasesPaginator(aa.client, accountAliasInput)
	return nil
}

func (aa *accountAliasMiner) generate(dummy string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := aa.fetchConf(&iam.ListAccountAliasesInput{}); err != nil {
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
