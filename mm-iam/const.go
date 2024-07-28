package main

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
	roleDetail          = "RoleDetail"
	roleInlinePolicy    = "RoleInlinePolicy"
	roleManagedPolicy   = "RoleManagedPolicy"
	roleInstanceProfile = "RoleInstanceProfile"

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
