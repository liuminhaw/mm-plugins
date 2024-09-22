package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/hashicorp/go-plugin"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

// Miner is a struct that implements the shared.Miner interface use in gRPC calls.
type Miner struct {
	resource shared.MinerResources
}

func (m Miner) Mine(mineConfig shared.MinerConfig) (shared.MinerResources, error) {
	log.Printf("Plugin name: %s\n", PLUG_NAME)

	// Get authentication profile from config
	awsAuth, err := utils.ConfigAuth(mineConfig)
	if err != nil {
		return nil, fmt.Errorf("mine: %w", err)
	}

	resources := shared.MinerResources{}
	for _, region := range awsAuth.Regions {
		log.Printf("Processing region: %s\n", region)

		cfg, err := config.LoadDefaultConfig(
			context.Background(),
			config.WithSharedConfigProfile(string(awsAuth.Profile)),
			config.WithRegion(region),
		)
		if err != nil {
			return nil, fmt.Errorf("mine: load config: %w", err)
		}

		client := sqs.NewFromConfig(cfg)
		paginator := sqs.NewListQueuesPaginator(
			client,
			&sqs.ListQueuesInput{MaxResults: aws.Int32(LIST_MAX_RESULTS)},
		)

		for paginator.HasMorePages() {
			fmt.Println("Processing next page")
			sqsOutput, err := paginator.NextPage(context.Background())
			if err != nil {
				return nil, fmt.Errorf("mine: list queues: %w", err)
			}

			for _, url := range sqsOutput.QueueUrls {
				log.Printf("Queue URL: %s\n", url)

				serviceClient := newSqsClient(client, url)
				sqsResource, err := utils.GetProperties(
					serviceClient,
					url,
					queueNameCache(url),
					propsConstructors,
				)
				if err != nil {
					var configErr *utils.MMError
					if errors.As(err, &configErr) {
						log.Printf("No properties in queue %s found", url)
						log.Printf("Error: %v\n", err)
					} else {
						log.Printf("mineResource: failed to get queue %s properties: %v", url, err)
					}
				} else {
					resources = append(resources, sqsResource)
				}
			}
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

// queueNameCache returns a CacheInfo struct with the queue name as the alias.
func queueNameCache(url string) utils.CacheInfo {
	parts := strings.Split(url, "/")

	return utils.CacheInfo{Alias: parts[len(parts)-1]}
}
