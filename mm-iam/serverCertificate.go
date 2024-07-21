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

type serverCertificateResource struct {
	client *iam.Client
}

func newServerCertificateResource(client *iam.Client) crawler {
	resource := serverCertificateResource{
		client: client,
	}
	return &resource
}

func (s *serverCertificateResource) fetchConf(input any) error {
	return nil
}

func (s *serverCertificateResource) generate(dummy cacheInfo) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: "ServerCertificate",
	}

	for _, prop := range miningServerCertificateProps {
		log.Printf("ServerCertificate property: %s\n", prop)

		serverCertificatePropsCrawler, err := newPropsCrawler(s.client, prop)
		if err != nil {
			return shared.MinerResource{}, fmt.Errorf("generate serverCertificateResource: %w", err)
		}
		serverCertificateProps, err := serverCertificatePropsCrawler.generate(dummy)
		if err != nil {
			var configErr *mmIAMError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", prop)
			} else {
				return shared.MinerResource{}, fmt.Errorf("generate serverCertificateResource: %w", err)
			}
		} else {
			resource.Properties = append(resource.Properties, serverCertificateProps...)
		}
	}

	// Check if there are any properties
	if resource.Properties == nil || len(resource.Properties) == 0 {
		return shared.MinerResource{}, &mmIAMError{"SSOProviders", noProps}
	}

	return resource, nil
}

// ServerCertificate detail
type serverCertificateDetailMiner struct {
	client        *iam.Client
	paginator     *iam.ListServerCertificatesPaginator
	configuration *iam.GetServerCertificateOutput
}

func newServerCertificateDetailMiner(client *iam.Client) propsCrawler {
	resource := serverCertificateDetailMiner{
		client: client,
	}
	return &resource
}

func (sc *serverCertificateDetailMiner) fetchConf(input any) error {
	serverCertificateInput, ok := input.(*iam.ListServerCertificatesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListServerCertificateInput type assertion failed")
	}

	sc.paginator = iam.NewListServerCertificatesPaginator(sc.client, serverCertificateInput)
	return nil
}

func (sc *serverCertificateDetailMiner) generate(dummy cacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := sc.fetchConf(&iam.ListServerCertificatesInput{}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate serverCertificate: %w", err)
	}

	for sc.paginator.HasMorePages() {
		page, err := sc.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate serverCertificate: %w", err)
		}

		for _, cert := range page.ServerCertificateMetadataList {
			sc.configuration, err = sc.client.GetServerCertificate(
				context.Background(),
				&iam.GetServerCertificateInput{
					ServerCertificateName: cert.ServerCertificateName,
				},
			)
			if err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate serverCertificate: %w", err)
			}

			property := shared.MinerProperty{
				Type: serverCertificateDetail,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(cert.ServerCertificateId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(sc.configuration.ServerCertificate); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate serverCertificate: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}
