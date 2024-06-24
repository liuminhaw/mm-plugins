package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
	"github.com/liuminhaw/mist-miner/shared"
)

var miningUserResoures = []string{
	userDetail,
	userLoginProfile,
	userAccessKey,
}

type userResource struct {
	client *iam.Client
}

func (u *userResource) fetchConf(input any) error {
	return nil
}

func (u *userResource) generate(mem *caching, idx int) (shared.MinerResource, error) {
	resource := shared.MinerResource{
		Identifier: fmt.Sprintf("User_%s", mem.users[idx].id),
	}

	for _, prop := range miningUserProps {
		log.Printf("user property: %s\n", prop)

		userPropsCrawler, err := newPropsCrawler(u.client, prop)
		if err != nil {
			return resource, fmt.Errorf("generate userResource: %w", err)
		}
		log.Printf("userPropsCrawler: %v\n", userPropsCrawler)
		userProps, err := userPropsCrawler.generate(mem.users[idx].name)
		if err != nil {
			var configErr *mmIAMError
			if errors.As(err, &configErr) {
				log.Printf("No %s configuration found", prop)
			} else {
				return resource, fmt.Errorf("generate userResource: %w", err)
			}
		} else {
			resource.Properties = append(resource.Properties, userProps...)
		}
	}

	return resource, nil
}

// user detail
type userDetailMiner struct {
	client        *iam.Client
	configuration *iam.GetUserOutput
}

func (ud *userDetailMiner) fetchConf(input any) error {
	userDetailInput, ok := input.(*iam.GetUserInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetUserInput type assertion failed")
	}

	var err error
	ud.configuration, err = ud.client.GetUser(context.Background(), userDetailInput)
	if err != nil {
		return fmt.Errorf("fetchConf userDetail: %w", err)
	}

	return nil
}

func (ud *userDetailMiner) generate(username string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := ud.fetchConf(&iam.GetUserInput{UserName: aws.String(username)}); err != nil {
		return properties, fmt.Errorf("generateUserDetail: %w", err)
	}

	property := shared.MinerProperty{
		Type: userDetail,
		Label: shared.MinerPropertyLabel{
			Name:   "UserDetail",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(ud.configuration.User); err != nil {
		return properties, fmt.Errorf("generate user detail: %w", err)
	}
	properties = append(properties, property)

	return properties, nil
}

// user login profile
type userLoginProfileMiner struct {
	client        *iam.Client
	configuration *iam.GetLoginProfileOutput
}

func (ulp *userLoginProfileMiner) fetchConf(input any) error {
	loginProfileInput, ok := input.(*iam.GetLoginProfileInput)
	if !ok {
		return fmt.Errorf("fetchConf: GetLoginProfileInput type assertion failed")
	}

	var err error
	ulp.configuration, err = ulp.client.GetLoginProfile(context.Background(), loginProfileInput)
	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "NoSuchEntity":
				return &mmIAMError{"LoginProfile", noConfig}
			default:
				return fmt.Errorf("fetchConf userLoginProfile: %w", err)
			}
		}
		return fmt.Errorf("fetchConf userLoginProfile: %w", err)
	}

	return nil
}

func (ulp *userLoginProfileMiner) generate(username string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := ulp.fetchConf(&iam.GetLoginProfileInput{UserName: aws.String(username)}); err != nil {
		return properties, fmt.Errorf("generate userLoginProfile: %w", err)
	}

	property := shared.MinerProperty{
		Type: userLoginProfile,
		Label: shared.MinerPropertyLabel{
			Name:   "LoginProfile",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatJson,
		},
	}
	if err := property.FormatContentValue(ulp.configuration.LoginProfile); err != nil {
		return properties, fmt.Errorf("generate userLoginProfile: %w", err)
	}
	properties = append(properties, property)

	return properties, nil
}

type userAccessKeyMiner struct {
	client    *iam.Client
	paginator *iam.ListAccessKeysPaginator
}

func (uak *userAccessKeyMiner) fetchConf(input any) error {
	listAccessKeysInput, ok := input.(*iam.ListAccessKeysInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAccessKeysInput type assertion failed")
	}

	uak.paginator = iam.NewListAccessKeysPaginator(uak.client, listAccessKeysInput)
	return nil
}

func (uak *userAccessKeyMiner) generate(username string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := uak.fetchConf(&iam.ListAccessKeysInput{UserName: aws.String(username)}); err != nil {
		return properties, fmt.Errorf("generate userAccessKey: %w", err)
	}

	for uak.paginator.HasMorePages() {
		page, err := uak.paginator.NextPage(context.Background())
		if err != nil {
			return properties, fmt.Errorf("generate user access key: %w", err)
		}

		for _, accessKey := range page.AccessKeyMetadata {
			property := shared.MinerProperty{
				Type: userAccessKey,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(accessKey.AccessKeyId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(accessKey); err != nil {
				return properties, fmt.Errorf("generate user access key: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}
