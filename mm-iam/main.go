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

	for _, resourceType := range miningResources {
		log.Printf("reosource type: %s\n", resourceType)
		resourceCrawler, err := New(client, resourceType)
		if err != nil {
			return nil, fmt.Errorf("Failed to create new crawler: %w", err)
		}
		log.Println("before generate")
		resource, err := resourceCrawler.generate()
		log.Println("after generate")
		if err != nil {
			log.Printf("Failed to get %s properties: %v", resourceType, err)
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
