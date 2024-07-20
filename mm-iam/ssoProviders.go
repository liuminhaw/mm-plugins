package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type ssoProvidersResource struct {
	client *iam.Client
}

func newSSOProvidersResource(client *iam.Client) crawler {
	resource := ssoProvidersResource{
		client: client,
	}
	return &resource
}

func (s *ssoProvidersResource) fetchConf(input any) error {
	return nil
}

func (s *ssoProvidersResource) generate(dummy cacheInfo) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: "SSOProviders",
	}

	for _, prop := range miningSSOProps {
		log.Printf("SSOProviders property: %s\n", prop)

		ssoPropsCrawler, err := newPropsCrawler(s.client, prop)
		if err != nil {
			return shared.MinerResource{}, fmt.Errorf("generate ssoResource: %w", err)
		}
		ssoProps, err := ssoPropsCrawler.generate("")
		if err != nil {
			var configErr *mmIAMError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", prop)
			} else {
				return shared.MinerResource{}, fmt.Errorf("generate ssoResource: %w", err)
			}
		} else {
			resource.Properties = append(resource.Properties, ssoProps...)
		}
	}

	// Check if there are any properties
	if resource.Properties == nil || len(resource.Properties) == 0 {
		return shared.MinerResource{}, &mmIAMError{"SSOProviders", noProps}
	}

	return resource, nil
}

// SSO OpenIDConnect provider
type ssoOIDCProviderMiner struct {
	client        *iam.Client
	configuration *iam.GetOpenIDConnectProviderOutput
	overview      *iam.ListOpenIDConnectProvidersOutput
}

func newSSOOIDCProviderMiner(client *iam.Client) *ssoOIDCProviderMiner {
	return &ssoOIDCProviderMiner{
		client: client,
	}
}

func (op *ssoOIDCProviderMiner) fetchConf(input any) error {
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

func (op *ssoOIDCProviderMiner) generate(dummy string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := op.fetchConf(""); err != nil {
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
	client        *iam.Client
	configuration *iam.GetSAMLProviderOutput
	overview      *iam.ListSAMLProvidersOutput
}

func newSSOSAMLProviderMiner(client *iam.Client) *ssoSAMLProviderMiner {
	return &ssoSAMLProviderMiner{
		client: client,
	}
}

func (sp *ssoSAMLProviderMiner) fetchConf(input any) error {
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

func (sp *ssoSAMLProviderMiner) generate(dummy string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := sp.fetchConf(""); err != nil {
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
