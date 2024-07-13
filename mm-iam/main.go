package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/go-plugin"
	"github.com/liuminhaw/mist-miner/shared"
	iamContext "github.com/liuminhaw/mm-plugins/mm-iam/context"
)

var PLUG_NAME = "mm-iam"

// This is the implementation of Miner
type Miner struct {
	resources shared.MinerResources
}

func (m Miner) Mine(mineConfig shared.MinerConfig) (shared.MinerResources, error) {
	log.Printf("Plugin name: %s\n", PLUG_NAME)

	// Get authentication profile from config
	awsAuth, err := configAuth(mineConfig)
	if err != nil {
		return nil, fmt.Errorf("mine: %w", err)
	}

	resources := shared.MinerResources{}
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(string(awsAuth.profile)),
	)

	client := iam.NewFromConfig(cfg)

	ctx := context.Background()
	if mineConfig.Equipments != nil {
		ctx = iamContext.WithEquipments(ctx, mineConfig.Equipments)
	}

	memory := newCaching()
	if err := memory.read(ctx, client); err != nil {
		return nil, fmt.Errorf("mine: %w", err)
	}

	for _, resourceType := range miningResources {
		log.Printf("resource type: %s\n", resourceType)

		switch resourceType {
		case iamUser:
			minedResources, err := mineResources(ctx, client, resourceType, memory.users)
			if err != nil {
				return nil, fmt.Errorf("mine: %w", err)
			}
			resources = append(resources, minedResources...)
		case iamGroup:
			minedResources, err := mineResources(ctx, client, resourceType, memory.groups)
			if err != nil {
				return nil, fmt.Errorf("mine: %w", err)
			}
			resources = append(resources, minedResources...)
		case iamPolicy:
			minedResources, err := mineResources(ctx, client, resourceType, memory.policies)
			if err != nil {
				return nil, fmt.Errorf("mine: %w", err)
			}
			resources = append(resources, minedResources...)
		case iamRole:
			minedResources, err := mineResources(ctx, client, resourceType, memory.roles)
			if err != nil {
				return nil, fmt.Errorf("mine: %w", err)
			}
			resources = append(resources, minedResources...)
		default:
			log.Printf("resource type: %s\n", resourceType)
		}
	}

	return resources, nil
}

func mineResources(
	ctx context.Context,
	client *iam.Client,
	resourceType string,
	data dataCache,
) (shared.MinerResources, error) {
	resources := shared.MinerResources{}
	for _, cache := range data.caches {
		log.Printf("Get %s: %s", data.resource, cache.name)
		resourceCrawler, err := NewCrawler(ctx, client, resourceType)
		if err != nil {
			return nil, fmt.Errorf("mineResource: failed to create new crawler: %w", err)
		}
		resource, err := resourceCrawler.generate(cache)
		if err != nil {
			log.Printf("mineResource: failed to get %s properties: %v", resourceType, err)
		} else {
			resources = append(resources, resource)
		}
	}

	return resources, nil
}

func main() {
	// logger setup for plugin logs
	log.SetOutput(os.Stderr)
	log.Printf("Starting miner plugin: %s\n", PLUG_NAME)

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"miner_grpc": &shared.MinerGRPCPlugin{Impl: &Miner{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

type awsProfile struct {
	profile string
}

// configAuth gets the authentication profile from the config
// and returns an awsProfile struct with the profile name
// for use in authenticating with AWS.
func configAuth(mineConfig shared.MinerConfig) (awsProfile, error) {
	if _, ok := mineConfig.Auth["profile"]; !ok {
		return awsProfile{}, fmt.Errorf("configAuth: profile not found")
	}

	return awsProfile{profile: mineConfig.Auth["profile"]}, nil
}
