package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/go-plugin"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

var PLUG_NAME = "mm-s3"

// This is the implementation of Miner
type Miner struct {
	resources shared.MinerResources
}

func (m Miner) Mine(mineConfig shared.MinerConfig) (shared.MinerResources, error) {
	log.Printf("Plugin name: %s\n", PLUG_NAME)

	// Get authentication profile from config
	awsAuth, err := utils.ConfigAuth(mineConfig)
	if err != nil {
		return nil, fmt.Errorf("mine: %w", err)
	}

	resources := shared.MinerResources{}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(string(awsAuth.Profile)),
	)

	client := s3.NewFromConfig(cfg)
	bucketsOutput, err := client.ListBuckets(context.Background(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("mine: list buckets: %w", err)
	}

	for _, bucket := range bucketsOutput.Buckets {
		log.Printf("Bucket: %s\n", *bucket.Name)

		bucketRegion, err := getBucketRegion(client, *bucket.Name)
		if err != nil {
			log.Printf("Failed to get bucket region: %v", err)
			continue
		}

		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithSharedConfigProfile(string(awsAuth.Profile)),
			config.WithRegion(bucketRegion),
		)

		serviceClient := newS3Client(s3.NewFromConfig(cfg), &bucket)
		bucketResource, err := utils.GetProperties(
			serviceClient,
			aws.ToString(bucket.Name),
			utils.CacheInfo{Name: location, Id: aws.ToString(bucket.Name), Content: bucketRegion},
			propsConstructors,
		)
		if err != nil {
			var configErr *utils.MMError
			if errors.As(err, &configErr) {
				log.Printf("No properties in bucket %s found", aws.ToString(bucket.Name))
			} else {
				log.Printf("mineResource: failed to get bucket %s properties: %v", aws.ToString(bucket.Name), err)
			}
		} else {
			bucketResource.Sort()
			resources = append(resources, bucketResource)
		}

	}

	return resources, nil
}

func main() {
	// logger setup for plugin logs
	log.SetOutput(os.Stderr)
	log.Println("Starting miner plugin")

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"miner_grpc": &shared.MinerGRPCPlugin{Impl: &Miner{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	},
	)
}

// getBucketRegion returns the region of the bucket
func getBucketRegion(client *s3.Client, bucket string) (string, error) {
	result, err := client.GetBucketLocation(context.Background(), &s3.GetBucketLocationInput{
		Bucket: &bucket,
	})
	if err != nil {
		return "", fmt.Errorf("getBucketRegion: %w", err)
	}

	region := string(result.LocationConstraint)
	if region == "" {
		region = "us-east-1"
	}

	return region, nil
}
