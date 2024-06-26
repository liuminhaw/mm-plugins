package main

import "fmt"

const (
	// users
	userDetail       = "UserDetail"
	userLoginProfile = "UserLoginProfile"
	userAccessKey    = "UserAccessKey"
	userMFADevice    = "UserMFADevice"

	accountAlias       = "AccountAlias"
	accessKey          = "AccessKey"
	iamGroup           = "Groups"
	iamInstanceProfile = "InstanceProfiles"
	iamUser            = "Users"
	mfaDevice          = "MFADevices"

	noConfig = "NoConfiguration"
)

var miningResources = []string{
	iamUser,
	// iamGroup,
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
}

type mmIAMError struct {
	category string
	code     string
}

func (e *mmIAMError) Error() string {
	return fmt.Sprintf("%s: %s", e.category, e.code)
}
