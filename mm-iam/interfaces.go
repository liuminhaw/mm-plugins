package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type crawler interface {
	fetchConf() error
	generate() (shared.MinerResource, error)
}

type propConstructor func(client *iam.Client) crawler

var propConstructors = map[string]propConstructor{
	iamUser: func(client *iam.Client) crawler {
		return &userResource{
			client: client,
		}
	},
	accessKey: func(client *iam.Client) crawler {
		return &accessKeyResource{
			client: client,
		}
	},
}

func New(client *iam.Client, propType string) (crawler, error) {
	constructor, ok := propConstructors[propType]
	if !ok {
		return nil, fmt.Errorf("New crawler: unknown property type: %s", propType)
	}
	return constructor(client), nil
}
