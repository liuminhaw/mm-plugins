package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type accessKeyResource struct {
	client    *iam.Client
	paginator *iam.ListAccessKeysPaginator
}

func (ar *accessKeyResource) fetchConf(input any) error {
	accessKeyInput, ok := input.(*iam.ListAccessKeysInput)
	if !ok {
		return fmt.Errorf("fetchConf: type assertion to *iam.ListAccessKeysInput failed")
	}
	ar.paginator = iam.NewListAccessKeysPaginator(ar.client, accessKeyInput)

	return nil
}

func (ar *accessKeyResource) generate(mem *caching) (shared.MinerResource, error) {
	resource := shared.MinerResource{Identifier: accessKey}

	if mem.usernames == nil {
		return resource, fmt.Errorf("generate accessKeyResource: caching usernames not found")
	}

	for _, username := range mem.usernames {
		log.Printf("Get key from user: %s\n", username)
		if err := ar.fetchConf(&iam.ListAccessKeysInput{UserName: aws.String(username)}); err != nil {
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

	}

	return resource, nil
}
