package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type ssoProvidersResource struct {
	client *iam.Client
}

func newSSOProvidersResource(client *iam.Client) utils.Crawler {
	resource := ssoProvidersResource{
		client: client,
	}
	return &resource
}

func (s *ssoProvidersResource) FetchConf(input any) error {
	return nil
}

func (s *ssoProvidersResource) Generate(dummy utils.CacheInfo) (shared.MinerResource, error) {
	return utils.GetProperties(
		s.client,
		"SSOProviders",
		dummy,
		ssoProvidersPropsCrawlerConstructors,
	)
}

// SSO OpenIDConnect provider
type ssoOIDCProviderMiner struct {
	propertyType  string
	client        *iam.Client
	configuration *iam.GetOpenIDConnectProviderOutput
	overview      *iam.ListOpenIDConnectProvidersOutput
}

func newSSOOIDCProviderMiner(client *iam.Client) *ssoOIDCProviderMiner {
	return &ssoOIDCProviderMiner{
		propertyType: ssoOIDCProvider,
		client:       client,
	}
}

func (op *ssoOIDCProviderMiner) PropertyType() string { return op.propertyType }

func (op *ssoOIDCProviderMiner) FetchConf(input any) error {
	var err error
	op.overview, err = op.client.ListOpenIDConnectProviders(
		context.Background(),
		&iam.ListOpenIDConnectProvidersInput{},
	)
	if err != nil {
		return fmt.Errorf("fetchConf SSO OIDC provider: %w", err)
	}

	return nil
}

func (op *ssoOIDCProviderMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := op.FetchConf(""); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate SSO OIDC provider: %w", err)
	}

	for _, provider := range op.overview.OpenIDConnectProviderList {
		output, err := op.client.GetOpenIDConnectProvider(
			context.Background(),
			&iam.GetOpenIDConnectProviderInput{
				OpenIDConnectProviderArn: provider.Arn,
			},
		)
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate SSO OIDC provider: %w", err)
		}

		property := shared.MinerProperty{
			Type: ssoOIDCProvider,
			Label: shared.MinerPropertyLabel{
				Name:   aws.ToString(provider.Arn),
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(output); err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate SSO OIDC provider: %w", err)
		}
		properties = append(properties, property)
	}

	return properties, nil
}

// SSO SAML provider
type ssoSAMLProviderMiner struct {
	propertyType  string
	client        *iam.Client
	configuration *iam.GetSAMLProviderOutput
	overview      *iam.ListSAMLProvidersOutput
}

func newSSOSAMLProviderMiner(client *iam.Client) *ssoSAMLProviderMiner {
	return &ssoSAMLProviderMiner{
		propertyType: ssoSAMLProvider,
		client:       client,
	}
}

func (sp *ssoSAMLProviderMiner) PropertyType() string { return sp.propertyType }

func (sp *ssoSAMLProviderMiner) FetchConf(input any) error {
	var err error
	sp.overview, err = sp.client.ListSAMLProviders(
		context.Background(),
		&iam.ListSAMLProvidersInput{},
	)
	if err != nil {
		return fmt.Errorf("fetchConf SSO SAML provider: %w", err)
	}

	return nil
}

func (sp *ssoSAMLProviderMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := sp.FetchConf(""); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate SSO SAML provider: %w", err)
	}

	for _, provider := range sp.overview.SAMLProviderList {
		output, err := sp.client.GetSAMLProvider(
			context.Background(),
			&iam.GetSAMLProviderInput{
				SAMLProviderArn: provider.Arn,
			},
		)
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate SSO SAML provider: %w", err)
		}

		property := shared.MinerProperty{
			Type: ssoSAMLProvider,
			Label: shared.MinerPropertyLabel{
				Name:   aws.ToString(provider.Arn),
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(output); err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate SSO SAML provider: %w", err)
		}
		properties = append(properties, property)
	}

	return properties, nil
}
