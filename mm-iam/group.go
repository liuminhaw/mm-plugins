package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
)

type groupResource struct {
	client *iam.Client
}

func (g *groupResource) fetchConf(input any) error {
	return nil
}

func (g *groupResource) generate(mem *caching, idx int) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: fmt.Sprintf("Group_%s", mem.groups[idx].id),
	}

	for _, prop := range miningGroupProps {
		log.Printf("group property: %s\n", prop)

		groupPropsCrawler, err := newPropsCrawler(g.client, prop)
		if err != nil {
			return resource, fmt.Errorf("generate groupResource: %w", err)
		}
		groupProps, err := groupPropsCrawler.generate(mem.groups[idx].name)
		if err != nil {
			var configErr *mmIAMError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", prop)
			} else {
				return resource, fmt.Errorf("generate groupResource: %w", err)
			}
		} else {
			resource.Properties = append(resource.Properties, groupProps...)
		}
	}

	return resource, nil
}

// group detail
type groupDetailMiner struct {
	client        *iam.Client
	configuration *iam.GetGroupOutput
}

func (gd *groupDetailMiner) fetchConf(input any) error {
	groupDetailInput, ok := input.(*iam.GetGroupInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetGroupInput type assertion failed")
	}

	var err error
	gd.configuration, err = gd.client.GetGroup(context.Background(), groupDetailInput)
	if err != nil {
		return fmt.Errorf("fetchConf groupDetail: %w", err)
	}

	return nil
}

func (gd *groupDetailMiner) generate(groupName string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := gd.fetchConf(&iam.GetGroupInput{GroupName: aws.String(groupName)}); err != nil {
		return properties, fmt.Errorf("generate groupDetail: %w", err)
	}

	// group detail
	property, err := gd.detailProp()
	if err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate groupDetail: %w", err)
	}
	properties = append(properties, property)

	// group users
	userProperties, err := gd.groupUserProp()
	if err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate groupDetail: %w", err)
	}
	properties = append(properties, userProperties...)

	return properties, nil
}

func (gd *groupDetailMiner) detailProp() (shared.MinerProperty, error) {
	property := shared.MinerProperty{
		Type: groupDetail,
		Label: shared.MinerPropertyLabel{
			Name:   "GroupDetail",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(gd.configuration.Group); err != nil {
		return shared.MinerProperty{}, fmt.Errorf("detailProp: %w", err)
	}

	return property, nil
}

func (gd *groupDetailMiner) groupUserProp() ([]shared.MinerProperty, error) {
	type groupUserInfo struct {
		Name string `json:"name"`
		Id   string `json:"id"`
	}

	properties := []shared.MinerProperty{}
	for _, user := range gd.configuration.Users {
		property := shared.MinerProperty{
			Type: groupUser,
			Label: shared.MinerPropertyLabel{
				Name:   "GroupUser",
				Unique: false,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(groupUserInfo{
			Name: aws.ToString(user.UserName),
			Id:   aws.ToString(user.UserId),
		}); err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate groupUser: %w", err)
		}
		properties = append(properties, property)
	}

	return properties, nil
}
