package main

import "fmt"

const (
	// users
	userDetail                    = "UserDetail"
	userLoginProfile              = "UserLoginProfile"
	userAccessKey                 = "UserAccessKey"
	userMFADevice                 = "UserMFADevice"
	userSSHPublicKey              = "UserSSHPublicKey"
	userServiceSpecificCredential = "UserServiceSpecificCredential"
	userSigningCertificate        = "UserSigningCertificate"

	// groups
	groupDetail = "GroupDetail"
	groupUser   = "GroupUser"

	// policies

	// crawlers
	iamGroup  = "Groups"
	iamUser   = "Users"
	iamPolicy = "Policies"

	noConfig = "NoConfiguration"

	// equipments
	policyEquipmentType = "policies"
)

var miningResources = []string{
	iamUser,
	iamGroup,
	iamPolicy,
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
	userMFADevice,
	userSSHPublicKey,
	userServiceSpecificCredential,
	userSigningCertificate,
}

var miningGroupProps = []string{
	groupDetail,
}

type mmIAMError struct {
	category string
	code     string
}

func (e *mmIAMError) Error() string {
	return fmt.Sprintf("%s: %s", e.category, e.code)
}
