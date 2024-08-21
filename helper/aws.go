package helper

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AwsDescribeRegionsApi interface {
	DescribeRegions(
		ctx context.Context,
		params *ec2.DescribeRegionsInput,
		optFns ...func(*ec2.Options),
	) (*ec2.DescribeRegionsOutput, error)
}

type AwsHelperAuth struct {
	Profile string
	Client  AwsDescribeRegionsApi
}

func NewAwsHelperAuth(profile string) (AwsHelperAuth, error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return AwsHelperAuth{}, fmt.Errorf("NewAwsHelperAuth(): %w", err)
	}

	client := ec2.NewFromConfig(cfg)

	return AwsHelperAuth{Profile: profile, Client: client}, nil
}

// AwsRegions returns a list of AWS regions.
// Set includeDisabled to true to include disabled regions (Get all regions).
func (auth AwsHelperAuth) AwsRegions(includeDisabled bool) ([]string, error) {
	output, err := auth.Client.DescribeRegions(
		context.Background(),
		&ec2.DescribeRegionsInput{AllRegions: aws.Bool(includeDisabled)},
	)
	if err != nil {
		return nil, fmt.Errorf("AwsRegions(%v): %w", includeDisabled, err)
	}

	regions := []string{}
	for _, region := range output.Regions {
		regions = append(regions, aws.ToString(region.RegionName))
	}

	return regions, nil
}
