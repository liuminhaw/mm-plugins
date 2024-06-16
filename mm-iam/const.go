package main

import "fmt"

const (
	iamUser = "Users"
)

var miningResources = []string{
	iamUser,
}

type mmIAMError struct {
	category string
	code     string
}

func (e *mmIAMError) Error() string {
	return fmt.Sprintf("%s: %s", e.category, e.code)
}
