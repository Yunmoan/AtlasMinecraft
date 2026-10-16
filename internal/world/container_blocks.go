package world

import (
	_ "embed"
	"fmt"
)

// BlockState.isRedstoneConductor is the original chest obstruction predicate;
// collision/solid-render guesses would incorrectly block transparent blocks.
//
//go:embed container_conductor_1.21.1.bin
var chestConductorFlags []byte

func chestRedstoneConductor(state uint32) (bool, error) {
	if len(chestConductorFlags) != (26684+7)/8 || state >= 26684 {
		return false, fmt.Errorf("original chest obstruction rules unavailable")
	}
	return chestConductorFlags[state>>3]&(1<<(state&7)) != 0, nil
}
