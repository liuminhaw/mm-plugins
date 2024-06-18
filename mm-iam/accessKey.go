package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type accessKeyResource struct {
	client    *iam.Client
	paginator *iam.ListAccessKeysPaginator
}

func (ar *accessKeyResource) fetchConf() error {
	ar.paginator = iam.NewListAccessKeysPaginator(ar.client, &iam.ListAccessKeysInput{})

	return nil
}

func (ar *accessKeyResource) generate() (shared.MinerResource, error) {
	resource := shared.MinerResource{Identifier: accessKey}

	if err := ar.fetchConf(); err != nil {
		return resource, fmt.Errorf("generate accessKeyResource: %w", err)
	}
	for ar.paginator.HasMorePages() {
		page, err := ar.paginator.NextPage(context.Background())
		if err != nil {
			return resource, fmt.Errorf(
				"generate accessKeyResource: failed to list iam access keys, %w",
				err,
			)
		}

		for _, keyData := range page.AccessKeyMetadata {
			property := shared.MinerProperty{
				Type: accessKey,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(keyData.AccessKeyId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(keyData); err != nil {
				return resource, fmt.Errorf("generate accessKeyResource: %w", err)
			}

			resource.Properties = append(resource.Properties, property)
		}
	}

	return resource, nil
}
