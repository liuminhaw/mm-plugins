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
	client *iam.Client
}

func newServerCertificateResource(client *iam.Client) utils.Crawler {
	resource := serverCertificateResource{
		client: client,
	}
	return &resource
}

func (s *serverCertificateResource) FetchConf(input any) error {
	return nil
}

func (s *serverCertificateResource) Generate(dummy utils.CacheInfo) (shared.MinerResource, error) {
	return utils.GetProperties(
		s.client,
		"ServerCertificate",
		dummy,
		serverCertificatePropsCrawlerConstructors,
	)
}

// ServerCertificate detail
type serverCertificateDetailMiner struct {
	propertyType  string
	client        *iam.Client
	paginator     *iam.ListServerCertificatesPaginator
	configuration *iam.GetServerCertificateOutput
}

func newServerCertificateDetailMiner(client *iam.Client) *serverCertificateDetailMiner {
	resource := serverCertificateDetailMiner{
		propertyType: serverCertificateDetail,
		client:       client,
	}
	return &resource
}

func (sc *serverCertificateDetailMiner) PropertyType() string { return sc.propertyType }

func (sc *serverCertificateDetailMiner) FetchConf(input any) error {
	serverCertificateInput, ok := input.(*iam.ListServerCertificatesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListServerCertificateInput type assertion failed")
	}

	sc.paginator = iam.NewListServerCertificatesPaginator(sc.client, serverCertificateInput)
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
