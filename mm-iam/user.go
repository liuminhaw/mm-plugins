package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type userResource struct {
	client *iam.Client
}

func newUserResource(client *iam.Client) utils.Crawler {
	resource := userResource{
		client: client,
	}
	return &resource
}

func (u *userResource) FetchConf(input any) error {
	return nil
}

func (u *userResource) Generate(datum utils.CacheInfo) (shared.MinerResource, error) {
	identifier := fmt.Sprintf("User_%s", datum.Id)
	return utils.GetProperties(u.client, identifier, datum, userPropsCrawlerConstructors)
}

// user detail (GetUser)
type userDetailMiner struct {
	propertyType  string
	client        *iam.Client
	configuration *iam.GetUserOutput
}

func newUserDetailMiner(client *iam.Client) *userDetailMiner {
	return &userDetailMiner{
		propertyType: userDetail,
		client:       client,
	}
}

func (ud *userDetailMiner) PropertyType() string { return ud.propertyType }

func (ud *userDetailMiner) FetchConf(input any) error {
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

func (ud *userDetailMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := ud.FetchConf(&iam.GetUserInput{UserName: aws.String(datum.Name)}); err != nil {
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

// user login profile (GetLoginProfile)
type userLoginProfileMiner struct {
	propertyType  string
	client        *iam.Client
	configuration *iam.GetLoginProfileOutput
}

func newUserLoginProfileMiner(client *iam.Client) *userLoginProfileMiner {
	return &userLoginProfileMiner{
		propertyType: userLoginProfile,
		client:       client,
	}
}

func (ulp *userLoginProfileMiner) PropertyType() string { return ulp.propertyType }

func (ulp *userLoginProfileMiner) FetchConf(input any) error {
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

func (ulp *userLoginProfileMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := ulp.FetchConf(&iam.GetLoginProfileInput{UserName: aws.String(datum.Name)}); err != nil {
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

// user accesskey (NewListAccessKeysPaginator)
type userAccessKeyMiner struct {
	propertyType string
	client       *iam.Client
	paginator    *iam.ListAccessKeysPaginator
}

func newUserAccessKeyMiner(client *iam.Client) *userAccessKeyMiner {
	return &userAccessKeyMiner{
		propertyType: userAccessKey,
		client:       client,
	}
}

func (uak *userAccessKeyMiner) PropertyType() string { return uak.propertyType }

func (uak *userAccessKeyMiner) FetchConf(input any) error {
	listAccessKeysInput, ok := input.(*iam.ListAccessKeysInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAccessKeysInput type assertion failed")
	}

	uak.paginator = iam.NewListAccessKeysPaginator(uak.client, listAccessKeysInput)
	return nil
}

func (uak *userAccessKeyMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := uak.FetchConf(&iam.ListAccessKeysInput{UserName: aws.String(datum.Name)}); err != nil {
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

// user MFA device (NewListMFADevicesPaginator)
type userMFADeviceMiner struct {
	propertyType string
	client       *iam.Client
	paginator    *iam.ListMFADevicesPaginator
}

func newUserMFADeviceMiner(client *iam.Client) *userMFADeviceMiner {
	return &userMFADeviceMiner{
		propertyType: userMFADevice,
		client:       client,
	}
}

func (umd *userMFADeviceMiner) PropertyType() string { return umd.propertyType }

func (umd *userMFADeviceMiner) FetchConf(input any) error {
	listMFADevicesInput, ok := input.(*iam.ListMFADevicesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListMFADevicesInput type assertion failed")
	}

	umd.paginator = iam.NewListMFADevicesPaginator(umd.client, listMFADevicesInput)
	return nil
}

func (umd *userMFADeviceMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := umd.FetchConf(&iam.ListMFADevicesInput{UserName: aws.String(datum.Name)}); err != nil {
		return properties, fmt.Errorf("generate userMFADevice: %w", err)
	}

	for umd.paginator.HasMorePages() {
		page, err := umd.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate user MFADevice: %w", err)
		}

		for _, mfaDevice := range page.MFADevices {
			// var property shared.MinerProperty
			property := shared.MinerProperty{
				Type: userMFADevice,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(mfaDevice.SerialNumber),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}

			// Check device type (virtual or hardware)
			if strings.Contains(aws.ToString(mfaDevice.SerialNumber), "mfa/") {
				log.Printf(
					"device: %s, type: virtual MFA Device",
					aws.ToString(mfaDevice.SerialNumber),
				)
				if err = property.FormatContentValue(mfaDevice); err != nil {
					return []shared.MinerProperty{}, fmt.Errorf("generate user MFADevice: %w", err)
				}
			} else {
				log.Printf("device: %s, type: hardware MFA Device", aws.ToString(mfaDevice.SerialNumber))
				device, err := umd.client.GetMFADevice(
					context.Background(),
					&iam.GetMFADeviceInput{SerialNumber: mfaDevice.SerialNumber},
				)
				if err != nil {
					return []shared.MinerProperty{}, fmt.Errorf("generate user MFADevice: %w", err)
				}
				property.Label.Name = aws.ToString(device.SerialNumber)
				if err = property.FormatContentValue(device); err != nil {
					return []shared.MinerProperty{}, fmt.Errorf("generate user MFADevice: %w", err)
				}
			}

			properties = append(properties, property)
		}
	}

	return properties, nil
}

// user SSH public key (NewListSSHPublicKeysPaginator)
type userSSHPublicKeyMiner struct {
	propertyType string
	client       *iam.Client
	paginator    *iam.ListSSHPublicKeysPaginator
}

func newUserSSHPublicKeyMiner(client *iam.Client) *userSSHPublicKeyMiner {
	return &userSSHPublicKeyMiner{
		propertyType: userSSHPublicKey,
		client:       client,
	}
}

func (uspk *userSSHPublicKeyMiner) PropertyType() string { return uspk.propertyType }

func (uspk *userSSHPublicKeyMiner) FetchConf(input any) error {
	sshPulicKeyInput, ok := input.(*iam.ListSSHPublicKeysInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListSSHPublicKeysInput type assertion failed")
	}

	uspk.paginator = iam.NewListSSHPublicKeysPaginator(uspk.client, sshPulicKeyInput)
	return nil
}

func (uspk *userSSHPublicKeyMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := uspk.FetchConf(&iam.ListSSHPublicKeysInput{UserName: aws.String(datum.Name)}); err != nil {
		return properties, fmt.Errorf("generate userSSHPublicKey: %w", err)
	}

	for uspk.paginator.HasMorePages() {
		page, err := uspk.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate user SSHPublicKey: %w", err)
		}

		for _, keyMetadata := range page.SSHPublicKeys {
			output, err := uspk.client.GetSSHPublicKey(
				context.Background(),
				&iam.GetSSHPublicKeyInput{
					Encoding:       types.EncodingTypePem,
					SSHPublicKeyId: keyMetadata.SSHPublicKeyId,
					UserName:       aws.String(datum.Name),
				},
			)
			if err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate user SSHPublicKey: %w", err)
			}

			property := shared.MinerProperty{
				Type: userSSHPublicKey,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(keyMetadata.SSHPublicKeyId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err = property.FormatContentValue(output.SSHPublicKey); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate user SSHPublicKey: %w", err)
			}

			properties = append(properties, property)
		}
	}

	return properties, nil
}

// user Service Specific Credential (ListServiceSpecificCredentials)
type userServiceSpecificCredentialMiner struct {
	propertyType  string
	client        *iam.Client
	configuration *iam.ListServiceSpecificCredentialsOutput
}

func newUserServiceSpecificCredentialMiner(client *iam.Client) *userServiceSpecificCredentialMiner {
	return &userServiceSpecificCredentialMiner{
		propertyType: userServiceSpecificCredential,
		client:       client,
	}
}

func (ussc *userServiceSpecificCredentialMiner) PropertyType() string { return ussc.propertyType }

func (ussc *userServiceSpecificCredentialMiner) FetchConf(input any) error {
	listServiceSpecificCredentialsInput, ok := input.(*iam.ListServiceSpecificCredentialsInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListServiceSpecificCredentialsInput type assertion failed")
	}

	var err error
	ussc.configuration, err = ussc.client.ListServiceSpecificCredentials(
		context.Background(),
		listServiceSpecificCredentialsInput,
	)
	if err != nil {
		return fmt.Errorf("fetchConf userServiceSpecificCredential: %w", err)
	}

	return nil
}

func (ussc *userServiceSpecificCredentialMiner) Generate(
	datum utils.CacheInfo,
) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := ussc.FetchConf(&iam.ListServiceSpecificCredentialsInput{UserName: aws.String(datum.Name)}); err != nil {
		return properties, fmt.Errorf("generate userServiceSpecificCredential: %w", err)
	}

	for _, credential := range ussc.configuration.ServiceSpecificCredentials {
		property := shared.MinerProperty{
			Type: userServiceSpecificCredential,
			Label: shared.MinerPropertyLabel{
				Name:   aws.ToString(credential.ServiceSpecificCredentialId),
				Unique: true,
			},
			Content: shared.MinerPropertyContent{
				Format: shared.FormatJson,
			},
		}
		if err := property.FormatContentValue(credential); err != nil {
			return properties, fmt.Errorf("generate user service specific credential: %w", err)
		}
		properties = append(properties, property)
	}

	return properties, nil
}

// user signing certificate (NewListSigningCertificatesPaginator)
type userSigningCertificateMiner struct {
	propertyType string
	client       *iam.Client
	paginator    *iam.ListSigningCertificatesPaginator
}

func newUserSigningCertificateMiner(client *iam.Client) *userSigningCertificateMiner {
	return &userSigningCertificateMiner{
		propertyType: userSigningCertificate,
		client:       client,
	}
}

func (usc *userSigningCertificateMiner) PropertyType() string { return usc.propertyType }

func (usc *userSigningCertificateMiner) FetchConf(input any) error {
	listSigningCertificatesInput, ok := input.(*iam.ListSigningCertificatesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListSigningCertificatesInput type assertion failed")
	}

	usc.paginator = iam.NewListSigningCertificatesPaginator(
		usc.client,
		listSigningCertificatesInput,
	)
	return nil
}

func (usc *userSigningCertificateMiner) Generate(
	datum utils.CacheInfo,
) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := usc.FetchConf(&iam.ListSigningCertificatesInput{UserName: aws.String(datum.Name)}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate userSigningCertificate: %w", err)
	}

	for usc.paginator.HasMorePages() {
		page, err := usc.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate user SigningCertificate: %w", err)
		}

		for _, certificate := range page.Certificates {
			property := shared.MinerProperty{
				Type: userSigningCertificate,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(certificate.CertificateId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(certificate); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf(
					"generate user SigningCertificate: %w",
					err,
				)
			}

			properties = append(properties, property)
		}
	}

	return properties, nil
}

// user inline policy
// Including information about the user's inline policies
type userInlinePolicyMiner struct {
	propertyType  string
	client        *iam.Client
	paginator     *iam.ListUserPoliciesPaginator
	configuration *iam.GetUserPolicyOutput
}

func newUserInlinePolicyMiner(client *iam.Client) *userInlinePolicyMiner {
	return &userInlinePolicyMiner{
		propertyType: userInlinePolicy,
		client:       client,
	}
}

func (uip *userInlinePolicyMiner) PropertyType() string { return uip.propertyType }

func (uip *userInlinePolicyMiner) FetchConf(input any) error {
	userInlinePolicyInput, ok := input.(*iam.ListUserPoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListUserPoliciesInput type assertion failed")
	}

	uip.paginator = iam.NewListUserPoliciesPaginator(uip.client, userInlinePolicyInput)
	return nil
}

func (uip *userInlinePolicyMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := uip.FetchConf(&iam.ListUserPoliciesInput{UserName: aws.String(datum.Name)}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate userInlinePolicy: %w", err)
	}

	for uip.paginator.HasMorePages() {
		page, err := uip.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate user InlinePolicy: %w", err)
		}

		for _, policyName := range page.PolicyNames {
			uip.configuration, err = uip.client.GetUserPolicy(
				context.Background(),
				&iam.GetUserPolicyInput{
					PolicyName: aws.String(policyName),
					UserName:   aws.String(datum.Name),
				},
			)
			if err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate user InlinePolicy: %w", err)
			}

			// Url decode on policy document
			decodedDocument, err := utils.DocumentUrlDecode(
				aws.ToString(uip.configuration.PolicyDocument),
			)
			if err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate user InlinePolicy: %w", err)
			}
			uip.configuration.PolicyDocument = aws.String(decodedDocument)

			property := shared.MinerProperty{
				Type: userInlinePolicy,
				Label: shared.MinerPropertyLabel{
					Name:   policyName,
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(uip.configuration); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate user InlinePolicy: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}

// user managed policy
// Including information about the user's managed policies
type userManagedPolicyMiner struct {
	propertyType string
	client       *iam.Client
	paginator    *iam.ListAttachedUserPoliciesPaginator
}

func newUserManagedPolicyMiner(client *iam.Client) *userManagedPolicyMiner {
	return &userManagedPolicyMiner{
		propertyType: userManagedPolicy,
		client:       client,
	}
}

func (ump *userManagedPolicyMiner) PropertyType() string { return ump.propertyType }

func (ump *userManagedPolicyMiner) FetchConf(input any) error {
	userManagedPolicyInput, ok := input.(*iam.ListAttachedUserPoliciesInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListAttachedUserPoliciesInput type assertion failed")
	}

	ump.paginator = iam.NewListAttachedUserPoliciesPaginator(ump.client, userManagedPolicyInput)
	return nil
}

func (ump *userManagedPolicyMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := ump.FetchConf(&iam.ListAttachedUserPoliciesInput{UserName: aws.String(datum.Name)}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate userManagedPolicy: %w", err)
	}

	for ump.paginator.HasMorePages() {
		page, err := ump.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate user ManagedPolicy: %w", err)
		}

		for _, policy := range page.AttachedPolicies {
			property := shared.MinerProperty{
				Type: userManagedPolicy,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(policy.PolicyName),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(policy); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate user ManagedPolicy: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}

// user belongs groups
type userGroupsMiner struct {
	propertyType string
	client       *iam.Client
	paginator    *iam.ListGroupsForUserPaginator
}

func newUserGroupsMiner(client *iam.Client) *userGroupsMiner {
	return &userGroupsMiner{
		propertyType: userGroups,
		client:       client,
	}
}

func (ug *userGroupsMiner) PropertyType() string { return ug.propertyType }

func (ug *userGroupsMiner) FetchConf(input any) error {
	userGroupsInput, ok := input.(*iam.ListGroupsForUserInput)
	if !ok {
		return fmt.Errorf("fetchConf: ListGroupsForUserInput type assertion failed")
	}

	ug.paginator = iam.NewListGroupsForUserPaginator(ug.client, userGroupsInput)
	return nil
}

func (ug *userGroupsMiner) Generate(datum utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	if err := ug.FetchConf(&iam.ListGroupsForUserInput{UserName: aws.String(datum.Name)}); err != nil {
		return []shared.MinerProperty{}, fmt.Errorf("generate userGroups: %w", err)
	}

	for ug.paginator.HasMorePages() {
		page, err := ug.paginator.NextPage(context.Background())
		if err != nil {
			return []shared.MinerProperty{}, fmt.Errorf("generate user Groups: %w", err)
		}

		for _, group := range page.Groups {
			property := shared.MinerProperty{
				Type: userGroups,
				Label: shared.MinerPropertyLabel{
					Name:   aws.ToString(group.GroupId),
					Unique: true,
				},
				Content: shared.MinerPropertyContent{
					Format: shared.FormatJson,
				},
			}
			if err := property.FormatContentValue(group); err != nil {
				return []shared.MinerProperty{}, fmt.Errorf("generate user Groups: %w", err)
			}
			properties = append(properties, property)
		}
	}

	return properties, nil
}
