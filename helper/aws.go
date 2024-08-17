package helper

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type AwsHelperAuth struct {
	Profile string
}

// AwsRegions returns a list of AWS regions.
// Set includeDisabled to true to include disabled regions (Get all regions).
func AwsRegions(auth AwsHelperAuth, includeDisabled bool) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile(string(auth.Profile)),
	)
	if err != nil {
		return nil, fmt.Errorf("AwsRegions(auth, %v): %w", includeDisabled, err)
	}

	client := ec2.NewFromConfig(cfg)
	output, err := client.DescribeRegions(
		context.Background(),
		&ec2.DescribeRegionsInput{AllRegions: aws.Bool(includeDisabled)},
	)
	if err != nil {
		return nil, fmt.Errorf("AwsRegions(auth, %v): %w", includeDisabled, err)
	}

	regions := []string{}
	for _, region := range output.Regions {
		regions = append(regions, aws.ToString(region.RegionName))
	}

	return regions, nil
}
