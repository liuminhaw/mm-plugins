package utils

import (
	"errors"
	"reflect"
	"testing"

	"github.com/liuminhaw/mist-miner/shared"
)

// TestConfigAuth_ValidConfig tests the ConfigAuth function with a valid configuration.
// TODO: Checking regions default setting if regions key is not found in config.
func TestConfigAuth(t *testing.T) {
	cases := map[string]struct {
		config shared.MinerConfig
		want   AwsProfile
		err    error
	}{
		"ValidConfig": {
			config: shared.MinerConfig{
				Auth: map[string]string{
					"profile": "testing",
					"regions": "us-west-1, us-west-2, ap-northeast-1",
				},
			},
			want: AwsProfile{
				Profile: "testing",
				Regions: []string{"us-west-1", "us-west-2", "ap-northeast-1"},
			},
		},
        "EmptyRegions": {
            config: shared.MinerConfig{
                Auth: map[string]string{
                    "profile": "testing",
                    "regions": "",
                },
            },
            want: AwsProfile{
                Profile: "testing",
                Regions: []string{},
            },
        },
		"MissingProfile": {
			config: shared.MinerConfig{},
			err:    errors.New("configAuth: profile not found"),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ConfigAuth(tc.config)
			if err != nil {
				if tc.err == nil {
					t.Fatalf("ConfigAuth() error: %v", err)
				}
				if err.Error() != tc.err.Error() {
					t.Errorf("ConfigAuth() error = %v; want %v", err, tc.err)
				}
			}

			if tc.err == nil && !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ConfigAuth() = %+v; want %+v", got, tc.want)
			}
		})
	}
}
