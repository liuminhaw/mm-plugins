package main

import "fmt"

const (
	iamUser   = "Users"
	accessKey = "AccessKey"
)

var miningResources = []string{
	iamUser,
	accessKey,
}

type mmIAMError struct {
	category string
	code     string
}

func (e *mmIAMError) Error() string {
	return fmt.Sprintf("%s: %s", e.category, e.code)
}
