package main

import (
	"context"

	"github.com/liuminhaw/mm-plugins/utils"
)

var crawlerConstructors = map[string]utils.CrawlerConstructor{
	iamUser: func(ctx context.Context, client utils.Client) (utils.Crawler, error) {
		return newUserResource(client)
	},
	iamGroup: func(ctx context.Context, client utils.Client) (utils.Crawler, error) {
		return newGroupResource(client)
	},
	iamPolicy: func(ctx context.Context, client utils.Client) (utils.Crawler, error) {
		return newPolicyResource(client)
	},
	iamRole: func(ctx context.Context, client utils.Client) (utils.Crawler, error) {
		return newRoleResource(client)
	},
	iamAccount: func(ctx context.Context, client utils.Client) (utils.Crawler, error) {
		return newAccountResource(client)
	},
	iamSSOProviders: func(ctx context.Context, client utils.Client) (utils.Crawler, error) {
		return newSSOProvidersResource(client)
	},
	iamServerCertificate: func(ctx context.Context, client utils.Client) (utils.Crawler, error) {
		return newServerCertificateResource(client)
	},
	iamVirtualMFADevice: func(ctx context.Context, client utils.Client) (utils.Crawler, error) {
		return newVirtualMFADeviceResource(client)
	},
	iamInstanceProfile: func(ctx context.Context, client utils.Client) (utils.Crawler, error) {
		return newInstanceProfileResource(client)
	},
}

// type propsCrawlerConstructor func(client *iam.Client) utils.PropsCrawler

var userPropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newUserDetailMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newUserLoginProfileMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newUserAccessKeyMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newUserMFADeviceMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newUserSSHPublicKeyMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newUserServiceSpecificCredentialMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newUserSigningCertificateMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newUserGroupsMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newUserInlinePolicyMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newUserManagedPolicyMiner(client)
	},
}

var groupPropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newGroupDetailMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newGroupInlinePolicyMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newGroupManagedPolicyMiner(client)
	},
}

var policyPropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newPolicyDetailMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newPolicyVersionsMiner(client)
	},
}

var rolePropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newRoleDetailMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newRoleInlinePolicyMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newRoleManagedPolicyMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newRoleInstanceProfileMiner(client)
	},
}

var accountPropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newAccountPasswordPolicyMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newAccountSummaryMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newAccountAliasMiner(client)
	},
}

var ssoProvidersPropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newSSOOIDCProviderMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newSSOSAMLProviderMiner(client)
	},
}

var serverCertificatePropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newServerCertificateDetailMiner(client)
	},
}

var virtualMFADevicePropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newVirtualMFADeviceDetailMiner(client)
	},
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newVirtualMFADeviceTagsMiner(client)
	},
}

var instanceProfilePropsCrawlerConstructors = []utils.PropsCrawlerConstructor{
	func(client utils.Client) (utils.PropsCrawler, error) {
		return newInstanceProfileDetailMiner(client)
	},
}
