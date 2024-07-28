package main

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mm-plugins/utils"
)

type iamClient struct {
	client *iam.Client
}

func newIAMClient(client *iam.Client) *iamClient {
	return &iamClient{client: client}
}

func (iamc *iamClient) Service() string { return "IAM" }

func assertIAMClient(serviceClient utils.Client) (*iamClient, error) {
	client, ok := serviceClient.(*iamClient)
	if !ok {
		return nil, errors.New("custom iamClient type assertion failed")
	}

	return client, nil
}
