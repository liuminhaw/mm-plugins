package main

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/liuminhaw/mm-plugins/utils"
)

type s3Client struct {
	client *s3.Client
	bucket *types.Bucket
}

func newS3Client(client *s3.Client, bucket *types.Bucket) *s3Client {
	return &s3Client{client: client, bucket: bucket}
}

// Implement the utils.Client interface
func (s3c *s3Client) Service() string { return "s3" }

func assertS3Client(serviceClient utils.Client) (*s3Client, error) {
	client, ok := serviceClient.(*s3Client)
	if !ok {
		return nil, errors.New("custom s3Client type assertion failed")
	}

	return client, nil
}
