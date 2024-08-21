package helper

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type mockAwsDescribeRegionsApi struct {
	describeRegionsFunc func(ctx context.Context, params *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error)
}

func (m mockAwsDescribeRegionsApi) DescribeRegions(
	ctx context.Context,
	input *ec2.DescribeRegionsInput,
	optFns ...func(*ec2.Options),
) (*ec2.DescribeRegionsOutput, error) {
	return m.describeRegionsFunc(ctx, input)
}

func TestAwsRegions(t *testing.T) {
	cases := map[string]struct {
		mockResponse *ec2.DescribeRegionsOutput
		want         []string
		err          error
	}{
		"Regions": {
			mockResponse: &ec2.DescribeRegionsOutput{
				Regions: []types.Region{
					{RegionName: aws.String("us-west-1")},
					{RegionName: aws.String("us-west-2")},
				},
			},
			want: []string{"us-west-1", "us-west-2"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockApi := mockAwsDescribeRegionsApi{
				describeRegionsFunc: func(ctx context.Context, params *ec2.DescribeRegionsInput) (*ec2.DescribeRegionsOutput, error) {
					return tc.mockResponse, nil
				},
			}

			auth := AwsHelperAuth{Profile: "testing", Client: mockApi}
			regions, err := auth.AwsRegions(false)
			if err != nil {
				t.Errorf("AwsRegions() error: %v", err)
			}

			if !equalStringSlices(regions, tc.want) {
				t.Errorf("AwsRegions() = %+v; want %+v", regions, tc.want)
			}
		})
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
