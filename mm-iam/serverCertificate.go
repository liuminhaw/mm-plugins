package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type serverCertificateResource struct {
	serviceClient *iamClient
}

func newServerCertificateResource(serviceClient utils.Client) (utils.Crawler, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newServerCertificateResource: %v", err)
	}

	return &serverCertificateResource{serviceClient: client}, nil
}

func (s *serverCertificateResource) FetchConf(input any) error {
	return nil
}

func (s *serverCertificateResource) Generate(dummy utils.CacheInfo) (shared.MinerResource, error) {
	return utils.GetProperties(
		s.serviceClient,
		"ServerCertificate",
		dummy,
		serverCertificatePropsCrawlerConstructors,
	)
}

// ServerCertificate detail
type serverCertificateDetailMiner struct {
	propertyType  string
	serviceClient *iamClient
	paginator     *iam.ListServerCertificatesPaginator
	configuration *iam.GetServerCertificateOutput
}

func newServerCertificateDetailMiner(
	serviceClient utils.Client,
) (*serverCertificateDetailMiner, error) {
	client, err := assertIAMClient(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newServerCertificateDetailMiner: %v", err)
	}

	return &serverCertificateDetailMiner{
		propertyType:  serverCertificateDetail,
		serviceClient: client,
	}, nil
}

func (sc *serverCertificateDetailMiner) PropertyType() string { return sc.propertyType }

func (sc *serverCertificateDetailMiner) FetchConf(input any) error {
	serverCertificateInput, ok := input.(*iam.ListServerCertificatesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListServerCertificateInput type assertion failed")
	}

	sc.paginator = iam.NewListServerCertificatesPaginator(
		sc.serviceClient.client,
		serverCertificateInput,
	)
	return nil
}

func (sc *serverCertificateDetailMiner) Generate(
	dummy utils.CacheInfo,
) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := sc.FetchConf(&iam.ListServerCertificatesInput{}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate serverCertificate: %w", err)
	}

	for sc.paginator.HasMorePages() {
		page, err := sc.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate serverCertificate: %w", err)
		}

		for _, cert := range page.ServerCertificateMetadataList {
			sc.configuration, err = sc.serviceClient.client.GetServerCertificate(
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
