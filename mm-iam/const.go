package main

import "fmt"

const (
	// users
	userDetail                    = "UserDetail"
	userLoginProfile              = "UserLoginProfile"
	userAccessKey                 = "UserAccessKey"
	userGroups                    = "UserGroups"
	userInlinePolicy              = "UserInlinePolicy"
	userMFADevice                 = "UserMFADevice"
	userManagedPolicy             = "UserManagedPolicy"
	userSSHPublicKey              = "UserSSHPublicKey"
	userServiceSpecificCredential = "UserServiceSpecificCredential"
	userSigningCertificate        = "UserSigningCertificate"

	// groups
	groupDetail        = "GroupDetail"
	groupUser          = "GroupUser"
	groupInlinePolicy  = "GroupInlinePolicy"
	groupManagedPolicy = "GroupManagedPolicy"

	// policies
	policyDetail   = "PolicyDetail"
	policyVersions = "PolicyVersions"

	// roles
	roleDetail        = "RoleDetail"
	roleInlinePolicy  = "RoleInlinePolicy"
	roleManagedPolicy = "RoleManagedPolicy"

	// Account
	accountPasswordPolicy = "AccountPasswordPolicy"
	accountSummary        = "AccountSummary"
	accountAlias          = "AccountAlias"

	// SSO Provider
	ssoOIDCProvider = "OIDCProvider"
	ssoSAMLProvider = "SAMLProvider"

	// crawlers
	iamGroup        = "Groups"
	iamUser         = "Users"
	iamPolicy       = "Policies"
	iamRole         = "Roles"
	iamAccount      = "Account"
	iamSSOProviders = "SSOProviders"

	noConfig = "NoConfiguration"
	noProps  = "NoProperties"

	// equipments
	policyEquipmentType = "policies"
)

var miningResources = []string{
	iamUser,
	iamGroup,
	iamPolicy,
	iamRole,
	iamAccount,
	iamSSOProviders,
	// iamInstanceProfile,
	// accountAlias,
	// acessKey will use username cache, should be placed after iamUser
	// accessKey,
	// mfaDevice will use username cache, should be placed after iamUser
	// mfaDevice,
}

var miningUserProps = []string{
	userDetail,
	userLoginProfile,
	userAccessKey,
	userGroups,
	userInlinePolicy,
	userManagedPolicy,
	userMFADevice,
	userSSHPublicKey,
	userServiceSpecificCredential,
	userSigningCertificate,
}

var miningGroupProps = []string{
	groupDetail,
	groupInlinePolicy,
	groupManagedPolicy,
}

var miningPolicyProps = []string{
	policyDetail,
	policyVersions,
}

var miningRoleProps = []string{
	roleDetail,
	roleInlinePolicy,
	roleManagedPolicy,
}

var miningAccountProps = []string{
	accountPasswordPolicy,
	accountSummary,
	accountAlias,
}

var miningSSOProps = []string{
	ssoOIDCProvider,
	ssoSAMLProvider,
}

type mmIAMError struct {
	category string
	code     string
}

func (e *mmIAMError) Error() string {
	return fmt.Sprintf("%s: %s", e.category, e.code)
}
