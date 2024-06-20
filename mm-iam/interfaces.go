package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type crawler interface {
	fetchConf(any) error
	generate(*caching) (shared.MinerResource, error)
}

type propConstructor func(client *iam.Client) crawler

var propConstructors = map[string]propConstructor{
	iamUser: func(client *iam.Client) crawler {
		return &userResource{
			client: client,
		}
	},
	iamGroup: func(client *iam.Client) crawler {
		return &groupResource{
			client: client,
		}
	},
    iamInstanceProfile: func(client *iam.Client) crawler {
        return &instanceProfileResource{
            client: client,
        }
    },
	accessKey: func(client *iam.Client) crawler {
		return &accessKeyResource{
			client: client,
		}
	},
	accountAlias: func(client *iam.Client) crawler {
		return &accountAliasResource{
			client: client,
		}
	},
    mfaDevice: func(client *iam.Client) crawler {
        return &mfaDeviceResource{
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
