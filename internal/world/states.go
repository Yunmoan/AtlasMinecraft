package world

import "atlas/internal/blockstate"

type BlockState = blockstate.BlockState

func BlockStateID(name string, properties map[string]string) (uint32, error) {
	return blockstate.BlockStateID(name, properties)
}

func BlockStateByID(id uint32) (BlockState, error) {
	return blockstate.BlockStateByID(id)
}
