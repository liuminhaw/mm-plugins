package main

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/liuminhaw/mm-plugins/utils"
)

type sqsClient struct {
	client *sqs.Client
	url    string
}

func newSqsClient(client *sqs.Client, queueUrl string) *sqsClient {
	return &sqsClient{client: client, url: queueUrl}
}

// Implement the utils.Client interface
func (sqs *sqsClient) Service() string { return "sqs" }

func assertSqsClient(serviceClient utils.Client) (*sqsClient, error) {
	client, ok := serviceClient.(*sqsClient)
	if !ok {
		return nil, errors.New("custom sqsClient type assertion failed")
	}

	return client, nil
}
