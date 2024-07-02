package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type crawler interface {
	fetchConf(any) error
	generate(*caching, int) (shared.MinerResource, error)
}

type crawlerConstructor func(client *iam.Client) crawler

var crawlerConstructors = map[string]crawlerConstructor{
	iamUser: func(client *iam.Client) crawler {
		return &userResource{
			client: client,
		}
	},
}

func NewCrawler(client *iam.Client, resourceType string) (crawler, error) {
	constructor, ok := crawlerConstructors[resourceType]
	if !ok {
		return nil, fmt.Errorf("New crawler: unknown property type: %s", resourceType)
	}
	return constructor(client), nil
}

type propsCrawler interface {
	fetchConf(any) error
	generate(string) ([]shared.MinerProperty, error)
}

type propsCrawlerConstructor func(client *iam.Client) propsCrawler

var propsCrawlerConstructors = map[string]propsCrawlerConstructor{
	userDetail: func(client *iam.Client) propsCrawler {
		return &userDetailMiner{
			client: client,
		}
	},
	userLoginProfile: func(client *iam.Client) propsCrawler {
		return &userLoginProfileMiner{
			client: client,
		}
	},
	userAccessKey: func(client *iam.Client) propsCrawler {
		return &userAccessKeyMiner{
			client: client,
		}
	},
	userMFADevice: func(client *iam.Client) propsCrawler {
		return &userMFADeviceMiner{
			client: client,
		}
	},
	userSSHPublicKey: func(client *iam.Client) propsCrawler {
		return &userSSHPublicKeyMiner{
			client: client,
		}
	},
	userServiceSpecificCredential: func(client *iam.Client) propsCrawler {
		return &userServiceSpecificCredentialMiner{
			client: client,
		}
	},
	userSigningCertificate: func(client *iam.Client) propsCrawler {
		return &userSigningCertificateMiner{
			client: client,
		}
	},
}

func newPropsCrawler(client *iam.Client, propType string) (propsCrawler, error) {
	constructor, ok := propsCrawlerConstructors[propType]
	if !ok {
		return nil, fmt.Errorf("New props crawler: unknown property type: %s", propType)
	}
	return constructor(client), nil
}
