package context

import (
	"context"

	"github.com/liuminhaw/mist-miner/shared"
)

type configKey string

const (
	configEquipmentsKey configKey = "equipments"
)

func WithEquipments(ctx context.Context, equipments []shared.MinerConfigEquipment) context.Context {
	return context.WithValue(ctx, configEquipmentsKey, equipments)
}

func Equipments(ctx context.Context) []shared.MinerConfigEquipment {
	val := ctx.Value(configEquipmentsKey)

	user, ok := val.([]shared.MinerConfigEquipment)
	if !ok {
		// The most likely case is that nothing was ever stored in the context,
		// so it doesn't have a type of []shared.MinerConfigEquipment. It is also possible that
		// other code in this package wrote an invalid value using the user key.
		return nil
	}

	return user
}
