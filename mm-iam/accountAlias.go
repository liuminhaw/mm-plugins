package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type accountAliasResource struct {
	client    *iam.Client
	paginator *iam.ListAccountAliasesPaginator
}

func (ar *accountAliasResource) fetchConf(input any) error {
	ar.paginator = iam.NewListAccountAliasesPaginator(ar.client, &iam.ListAccountAliasesInput{})

	return nil
}

func (ar *accountAliasResource) generate(memory *caching) (shared.MinerResource, error) {
	resource := shared.MinerResource{Identifier: accountAlias}

	if err := ar.fetchConf(nil); err != nil {
		return resource, fmt.Errorf("generate accountAliasResource: %w", err)
	}
	for ar.paginator.HasMorePages() {
		page, err := ar.paginator.NextPage(context.Background())
		if err != nil {
			return resource, fmt.Errorf(
				"generate accountAliasResource: failed to list account alias %w",
				err,
			)
		}

		for _, alias := range page.AccountAliases {
			property := shared.MinerProperty{
				Type: "Alias",
				Label: shared.MinerPropertyLabel{
					Name:   "Alias",
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatText,
				},
			}
			if err := property.FormatContentValue(alias); err != nil {
				return resource, fmt.Errorf("generate accountAliasResource: %w", err)
			}

			resource.Properties = append(resource.Properties, property)
		}
	}

	return resource, nil
}
