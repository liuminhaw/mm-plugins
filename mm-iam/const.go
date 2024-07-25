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

	// Server Certificate
	serverCertificateDetail = "ServerCertificateDetail"

	// Virtual MFA
	virtualMFADeviceDetail = "VirtualMFADeviceDetail"
	virtualMFADeviceTags   = "VirtualMFADeviceTags"

	// Instance Profile
	instanceProfileDetail = "InstanceProfileDetail"

	// crawlers
	iamGroup             = "Groups"
	iamUser              = "Users"
	iamPolicy            = "Policies"
	iamRole              = "Roles"
	iamAccount           = "Account"
	iamSSOProviders      = "SSOProviders"
	iamServerCertificate = "ServerCertificate"
	iamVirtualMFADevice  = "VirtualMFADevice"
	iamInstanceProfile   = "InstanceProfile"

	noConfig = "NoConfiguration"
	noProps  = "NoProperties"

	// equipments
	policyEquipmentType     = "policies"
	virtualMFAEquipmentType = "virtualMFADevices"
)

var miningResources = []string{
	iamUser,
	iamGroup,
	iamPolicy,
	iamRole,
	iamAccount,
	iamSSOProviders,
	iamServerCertificate,
	iamVirtualMFADevice,
	iamInstanceProfile,
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

var miningServerCertificateProps = []string{
	serverCertificateDetail,
}

var miningVirtualMFAProps = []string{
	virtualMFADeviceDetail,
	virtualMFADeviceTags,
}

var miningInstanceProfileProps = []string{
	instanceProfileDetail,
}

type mmIAMError struct {
	category string
	code     string
}

func (e *mmIAMError) Error() string {
	return fmt.Sprintf("%s: %s", e.category, e.code)
}
