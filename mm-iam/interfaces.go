package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type crawler interface {
	fetchConf(any) error
	generate(cacheInfo) (shared.MinerResource, error)
}

type crawlerConstructor func(ctx context.Context, client *iam.Client) crawler

var crawlerConstructors = map[string]crawlerConstructor{
	iamUser: func(ctx context.Context, client *iam.Client) crawler {
		return newUserResource(client)
	},
	iamGroup: func(ctx context.Context, client *iam.Client) crawler {
		return newGroupResource(client)
	},
	iamPolicy: func(ctx context.Context, client *iam.Client) crawler {
		return newPolicyResource(client)
	},
	iamRole: func(ctx context.Context, client *iam.Client) crawler {
		return newRoleResource(client)
	},
}

func NewCrawler(ctx context.Context, client *iam.Client, resourceType string) (crawler, error) {
	constructor, ok := crawlerConstructors[resourceType]
	if !ok {
		return nil, fmt.Errorf("New crawler: unknown property type: %s", resourceType)
	}
	return constructor(ctx, client), nil
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
	userInlinePolicy: func(client *iam.Client) propsCrawler {
		return newUserInlinePolicyMiner(client)
	},
	userManagedPolicy: func(client *iam.Client) propsCrawler {
		return newUserManagedPolicyMiner(client)
	},
	groupDetail: func(client *iam.Client) propsCrawler {
		return &groupDetailMiner{
			client: client,
		}
	},
	groupInlinePolicy: func(client *iam.Client) propsCrawler {
		return newGroupInlinePolicyMiner(client)
	},
	groupManagedPolicy: func(client *iam.Client) propsCrawler {
		return newGroupManagedPolicyMiner(client)
	},
	policyDetail: func(client *iam.Client) propsCrawler {
		return &policyDetailMiner{
			client: client,
		}
	},
	policyVersions: func(client *iam.Client) propsCrawler {
		return &policyVersionsMiner{
			client: client,
		}
	},
	roleDetail: func(client *iam.Client) propsCrawler {
		return &roleDetailMiner{
			client: client,
		}
	},
	roleInlinePolicy: func(client *iam.Client) propsCrawler {
		return &roleInlinePolicyMiner{
			client: client,
		}
	},
	roleManagedPolicy: func(client *iam.Client) propsCrawler {
		return &roleManagedPolicyMiner{
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
