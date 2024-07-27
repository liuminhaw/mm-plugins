package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mm-plugins/utils"
)

var crawlerConstructors = map[string]utils.CrawlerConstructor{
	iamUser: func(ctx context.Context, client *iam.Client) utils.Crawler {
		return newUserResource(client)
	},
	iamGroup: func(ctx context.Context, client *iam.Client) utils.Crawler {
		return newGroupResource(client)
	},
	iamPolicy: func(ctx context.Context, client *iam.Client) utils.Crawler {
		return newPolicyResource(client)
	},
	iamRole: func(ctx context.Context, client *iam.Client) utils.Crawler {
		return newRoleResource(client)
	},
	iamAccount: func(ctx context.Context, client *iam.Client) utils.Crawler {
		return newAccountResource(client)
	},
	iamSSOProviders: func(ctx context.Context, client *iam.Client) utils.Crawler {
		return newSSOProvidersResource(client)
	},
	iamServerCertificate: func(ctx context.Context, client *iam.Client) utils.Crawler {
		return newServerCertificateResource(client)
	},
	iamVirtualMFADevice: func(ctx context.Context, client *iam.Client) utils.Crawler {
		return newVirtualMFADeviceResource(client)
	},
	iamInstanceProfile: func(ctx context.Context, client *iam.Client) utils.Crawler {
		return newInstanceProfileResource(client)
	},
}

type propsCrawlerConstructor func(client *iam.Client) utils.PropsCrawler

var userPropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client *iam.Client) utils.PropsCrawler {
		return newUserDetailMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newUserLoginProfileMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newUserAccessKeyMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newUserMFADeviceMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newUserSSHPublicKeyMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newUserServiceSpecificCredentialMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newUserSigningCertificateMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newUserGroupsMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newUserInlinePolicyMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newUserManagedPolicyMiner(client)
	},
}

var groupPropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client *iam.Client) utils.PropsCrawler {
		return newGroupDetailMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newGroupInlinePolicyMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newGroupManagedPolicyMiner(client)
	},
}

var policyPropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client *iam.Client) utils.PropsCrawler {
		return newPolicyDetailMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newPolicyVersionsMiner(client)
	},
}

var rolePropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client *iam.Client) utils.PropsCrawler {
		return newRoleDetailMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newRoleInlinePolicyMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newRoleManagedPolicyMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newRoleInstanceProfileMiner(client)
	},
}

var accountPropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client *iam.Client) utils.PropsCrawler {
		return newAccountPasswordPolicyMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newAccountSummaryMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newAccountAliasMiner(client)
	},
}

var ssoProvidersPropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client *iam.Client) utils.PropsCrawler {
		return newSSOOIDCProviderMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newSSOSAMLProviderMiner(client)
	},
}

var serverCertificatePropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client *iam.Client) utils.PropsCrawler {
		return newServerCertificateDetailMiner(client)
	},
}

var virtualMFADevicePropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client *iam.Client) utils.PropsCrawler {
		return newVirtualMFADeviceDetailMiner(client)
	},
	func(client *iam.Client) utils.PropsCrawler {
		return newVirtualMFADeviceTagsMiner(client)
	},
}

var instanceProfilePropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client *iam.Client) utils.PropsCrawler {
		return newInstanceProfileDetailMiner(client)
	},
}
