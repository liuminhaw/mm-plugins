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
	// configuration *iam.GetUserOutput

	detail       *userDetailResource
	loginProfile *userLoginProfileResource
	accessKey    *userAccessKeyResource
}

func NewUserResource(client *iam.Client) *userResource {
	return &userResource{
		client:       client,
		detail:       &userDetailResource{},
		loginProfile: &userLoginProfileResource{},
		accessKey:    &userAccessKeyResource{},
	}
}

func (u *userResource) fetchConf(input any) error {
	var err error

	switch input.(type) {
	case *iam.GetUserInput:
		u.detail.configuration, err = u.client.GetUser(context.Background(), input.(*iam.GetUserInput))
	case *iam.GetLoginProfileInput:
		u.loginProfile.configuration, err = u.client.GetLoginProfile(context.Background(), input.(*iam.GetLoginProfileInput))
	case *iam.ListAccessKeysInput:
		u.accessKey.paginator = iam.NewListAccessKeysPaginator(u.client, input.(*iam.ListAccessKeysInput))
	default:
		return fmt.Errorf("fetchConf: unknown input type")
	}

	if err != nil {
		var apiErr smithy.APIError
		if ok := errors.As(err, &apiErr); ok {
			switch apiErr.ErrorCode() {
			case "NoSuchEntity":
				return &mmIAMError{"LoginProfile", noConfig}
			default:
				return fmt.Errorf("fetchConf LoginProfile: %w", err)
			}
		}
		return fmt.Errorf("fetchConf userResource: %w", err)
	}

	return nil
}

func (u *userResource) generate(username string) (shared.MinerResource, error) {
	resource := shared.MinerResource{}

	// user detail
	if err := u.fetchConf(&iam.GetUserInput{UserName: aws.String(username)}); err != nil {
		return shared.MinerResource{}, fmt.Errorf("generate userResource: %w", err)
	} else {
		property, err := u.detail.generate(username)
		if err != nil {
			return resource, fmt.Errorf("generate userResource: %w", err)
		}
		resource.Identifier = fmt.Sprintf("User_%s", aws.ToString(u.detail.configuration.User.UserId))
		resource.Properties = append(resource.Properties, property)
	}

	// login profile
	var configErr *mmIAMError
	if err := u.fetchConf(&iam.GetLoginProfileInput{UserName: aws.String(username)}); err != nil {
		if errors.As(err, &configErr) {
			log.Print("No loginProfile configuration found")
		} else {
			return shared.MinerResource{}, fmt.Errorf("generate userResource: %w", err)
		}
	} else {
		property, err := u.loginProfile.generate(username)
		if err != nil {
			return resource, fmt.Errorf("generate userResource: %w", err)
		}
		resource.Properties = append(resource.Properties, property)
	}

	// access key
	if err := u.fetchConf(&iam.ListAccessKeysInput{UserName: aws.String(username)}); err != nil {
		return shared.MinerResource{}, fmt.Errorf("generate userResource: %w", err)
	} else {
		properties, err := u.accessKey.generate(username)
		if err != nil {
			return resource, fmt.Errorf("generate userResource: %w", err)
		}
		resource.Properties = append(resource.Properties, properties...)
	}

	return resource, nil
}

type userCrawler interface {
	generate(username string) (shared.MinerResource, error)
}

// user detail
type userDetailResource struct {
	configuration *iam.GetUserOutput
}

func (ud *userDetailResource) generate(username string) (shared.MinerProperty, error) {
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
		return shared.MinerProperty{}, fmt.Errorf("generate user detail: %w", err)
	}

	return property, nil
}

// user login profile
type userLoginProfileResource struct {
	configuration *iam.GetLoginProfileOutput
}

func (ulp *userLoginProfileResource) generate(username string) (shared.MinerProperty, error) {
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
		return shared.MinerProperty{}, fmt.Errorf("generate user login profile: %w", err)
	}

	return property, nil
}

type userAccessKeyResource struct {
	paginator *iam.ListAccessKeysPaginator
}

func (uak *userAccessKeyResource) generate(username string) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

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
