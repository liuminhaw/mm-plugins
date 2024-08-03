package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type replicationMiner struct {
	propertyType  string
	serviceClient *s3Client
	configuration *s3.GetBucketReplicationOutput
}

func newReplicationMiner(serviceClient utils.Client, property string) (*replicationMiner, error) {
	client, err := assertS3Client(serviceClient)
	if err != nil {
		return nil, fmt.Errorf("newReplicationMiner: %w", err)
	}

	return &replicationMiner{propertyType: property, serviceClient: client}, nil
}

func (r *replicationMiner) PropertyType() string { return r.propertyType }

func (r *replicationMiner) FetchConf(input any) error {
	replicationInput, ok := input.(*s3.GetBucketReplicationInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetBucketReplicationInput type assertion failed")
	}

	var err error
	r.configuration, err = r.serviceClient.client.GetBucketReplication(
		context.Background(),
		replicationInput,
	)
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "ReplicationConfigurationNotFoundError":
				return &utils.MMError{Category: replication, Code: utils.NoConfig}
			default:
				return fmt.Errorf("fetchConf bucket replication: %w", err)
			}
		}
		return fmt.Errorf("fetchConf bucket replication: %w", err)
	}

	return nil
}

func (r *replicationMiner) Generate(dummy utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := r.FetchConf(&s3.GetBucketReplicationInput{Bucket: r.serviceClient.bucket.Name}); err != nil {
		return nil, fmt.Errorf("generate bucket replication: %w", err)
	}

	if r.configuration.ReplicationConfiguration != nil {
		property := shared.MinerProperty{
			Type: replication,
			Label: shared.MinerPropertyLabel{
				Name:   "Replication",
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(r.configuration.ReplicationConfiguration); err != nil {
			return nil, fmt.Errorf("generate buckeP replication: %w", err)
		}

		properties = append(properties, property)
	}

	return properties, nil
}
