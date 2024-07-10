package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/liuminhaw/mist-miner/shared"
	iamContext "github.com/liuminhaw/mm-plugins/mm-iam/context"
)

type policyResource struct {
	client     *iam.Client
	equipments []shared.MinerConfigEquipment
}

func newPolicyResource(ctx context.Context, client *iam.Client) crawler {
	resource := policyResource{
		client: client,
	}
	equipments := iamContext.Equipments(ctx)
	resource.readEquipments(equipments)

	return &resource
}

func (p *policyResource) fetchConf(input any) error {
	return nil
}

func (p *policyResource) generate(mem *caching, idx int) (shared.MinerResource, error) {
	for _, equipment := range p.equipments {
		fmt.Printf("Equipment type: %s\n", equipment.Type)
		fmt.Printf("Equipment Name: %s\n", equipment.Name)
		for key, value := range equipment.Attributes {
			fmt.Printf("Attribute Name: %s\n", key)
			fmt.Printf("Attribute Value: %s\n", value)
		}
	}

	return shared.MinerResource{}, nil
}

func (p *policyResource) readEquipments(equipments []shared.MinerConfigEquipment) {
	p.equipments = []shared.MinerConfigEquipment{}

	for _, equipment := range equipments {
		if equipment.Type == policyEquipmentType {
			p.equipments = append(p.equipments, equipment)
		}
	}
}
