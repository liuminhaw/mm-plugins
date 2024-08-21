package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/helper"
)

type AwsProfile struct {
	Profile string
	Regions []string
}

// configAuth gets the authentication profile from the config
// and returns an awsProfile struct with the profile name
// for use in authenticating with AWS.
func ConfigAuth(mineConfig shared.MinerConfig) (AwsProfile, error) {
	// Read profile setting from config
	if _, ok := mineConfig.Auth["profile"]; !ok {
		return AwsProfile{}, fmt.Errorf("configAuth: profile not found")
	}

	// Read regions setting from config if exists
	regions := []string{}
	if _, ok := mineConfig.Auth["regions"]; !ok {
		auth, err := helper.NewAwsHelperAuth(mineConfig.Auth["profile"])
		if err != nil {
			return AwsProfile{}, fmt.Errorf("configAuth: %w", err)
		}
		regions, err = auth.AwsRegions(false)
		if err != nil {
			return AwsProfile{}, fmt.Errorf("configAuth: %w", err)
		}
	} else {
		regionsList := strings.Split(mineConfig.Auth["regions"], ",")
		for _, region := range regionsList {
			regions = append(regions, strings.TrimSpace(region))
		}
	}

	return AwsProfile{Profile: mineConfig.Auth["profile"], Regions: regions}, nil
}

var ErrAttributeNotFound = errors.New("miner configuration attribute not found")

type EquipmentInfo struct {
	AcceptVals []string
	DefaultVal string
	TargetType string
	TargetName string
	TargetAttr string
}

// GetEquipAttribute read from given equipments and return the attribute value
// that matches the given EquipmentInfo.
// If the attribute is not found, return the default value in equipment info.
func GetEquipAttribute(
	equipments []shared.MinerConfigEquipment,
	info EquipmentInfo,
) string {
	var result string

	for _, equipment := range equipments {
		if equipment.Type == info.TargetType && equipment.Name == info.TargetName {
			result = equipment.Attributes[info.TargetAttr]
		}
	}

	for _, v := range info.AcceptVals {
		if result == v {
			return result
		}
	}
	return info.DefaultVal
}
