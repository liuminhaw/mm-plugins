package main

import (
	"github.com/liuminhaw/mist-miner/shared"
	"github.com/liuminhaw/mm-plugins/utils"
)

type regionMiner struct {
	propertyType string
}

func newRegionMiner(property string) (*regionMiner, error) {
	return &regionMiner{propertyType: property}, nil
}

func (r *regionMiner) PropertyType() string { return r.propertyType }

func (r *regionMiner) FetchConf(input any) error { return nil }

func (r *regionMiner) Generate(data utils.CacheInfo) ([]shared.MinerProperty, error) {
	properties := []shared.MinerProperty{}

	property := shared.MinerProperty{
		Type: location,
		Label: shared.MinerPropertyLabel{
			Name:   "Region",
			Unique: true,
		},
		Content: shared.MinerPropertyContent{
			Format: shared.FormatText,
			Value:  data.Content,
		},
	}
	properties = append(properties, property)

	return properties, nil
}
