package main

import "fmt"

const (
	accountAlias       = "AccountAlias"
	accessKey          = "AccessKey"
	iamGroup           = "Groups"
	iamInstanceProfile = "InstanceProfiles"
	iamUser            = "Users"
    mfaDevice          = "MFADevices"
)

var miningResources = []string{
	iamUser,
	iamGroup,
	iamInstanceProfile,
	accountAlias,
	// acessKey will use username cache, should be placed after iamUser
	accessKey,
    // mfaDevice will use username cache, should be placed after iamUser
    mfaDevice,
}

type mmIAMError struct {
	category string
	code     string
}

func (e *mmIAMError) Error() string {
	return fmt.Sprintf("%s: %s", e.category, e.code)
}
