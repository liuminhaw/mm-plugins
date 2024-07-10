package utils

import (
	"errors"

	"github.com/liuminhaw/mist-miner/shared"
)

var ErrAttributeNotFound = errors.New("miner configuration attribute not found")

func GetEquipAttribute(
	equipments []shared.MinerConfigEquipment,
	targetType, targetName, targetAttr string,
) (string, error) {
	var result string

	for _, equipment := range equipments {
		if equipment.Type == targetType && equipment.Name == targetName {
			result = equipment.Attributes[targetAttr]
		}
	}

	if result == "" {
		return "", ErrAttributeNotFound
	}
	return result, nil
}
