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
	serviceClient *iamClient
}

func newSSOProvidersResource(serviceClient utils.Client) (utils.Crawler, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newSSOProvidersResource: %v", err)
	}

	return &ssoProvidersResource{serviceClient: client}, nil
}

func (s *ssoProvidersResource) FetchConf(input any) error {
	return nil
}

func (s *ssoProvidersResource) Generate(dummy utils.CacheInfo) (shared.MinerResource, error) {
	return utils.GetProperties(
		s.serviceClient,
		"SSOProviders",
		dummy,
		ssoProvidersPropsCrawlerConstructors,
	)
}

// SSO OpenIDConnect provider
type ssoOIDCProviderMiner struct {
	propertyType  string
	serviceClient *iamClient
	configuration *iam.GetOpenIDConnectProviderOutput
	overview      *iam.ListOpenIDConnectProvidersOutput
}

func newSSOOIDCProviderMiner(serviceClient utils.Client) (*ssoOIDCProviderMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newSSOOIDCProviderMiner: %v", err)
	}

	return &ssoOIDCProviderMiner{
		propertyType:  ssoOIDCProvider,
		serviceClient: client,
	}, nil
}

func (op *ssoOIDCProviderMiner) PropertyType() string { return op.propertyType }

func (op *ssoOIDCProviderMiner) FetchConf(input any) error {
	var err error
	op.overview, err = op.serviceClient.client.ListOpenIDConnectProviders(
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
		output, err := op.serviceClient.client.GetOpenIDConnectProvider(
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
	serviceClient *iamClient
	configuration *iam.GetSAMLProviderOutput
	overview      *iam.ListSAMLProvidersOutput
}

func newSSOSAMLProviderMiner(serviceClient utils.Client) (*ssoSAMLProviderMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newSSOSAMLProviderMiner: %v", err)
	}

	return &ssoSAMLProviderMiner{
		propertyType:  ssoSAMLProvider,
		serviceClient: client,
	}, nil
}

func (sp *ssoSAMLProviderMiner) PropertyType() string { return sp.propertyType }

func (sp *ssoSAMLProviderMiner) FetchConf(input any) error {
	var err error
	sp.overview, err = sp.serviceClient.client.ListSAMLProviders(
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
		output, err := sp.serviceClient.client.GetSAMLProvider(
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
