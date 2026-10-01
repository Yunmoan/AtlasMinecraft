package world

import "testing"

func TestVanillaBlockStateTable(t *testing.T) {
	cases := []struct {
		name  string
		props map[string]string
		id    uint32
	}{
		{"minecraft:air", nil, 0},
		{"minecraft:grass_block", map[string]string{"snowy": "false"}, 9},
		{"minecraft:grass_block", nil, 9},
		{"minecraft:oak_log", map[string]string{"axis": "y"}, 131},
		{"minecraft:oak_stairs", map[string]string{"facing": "north", "half": "bottom", "shape": "straight", "waterlogged": "false"}, 2885},
		{"minecraft:oak_stairs", map[string]string{"facing": "north"}, 2885},
	}
	for _, tc := range cases {
		got, err := BlockStateID(tc.name, tc.props)
		if err != nil || got != tc.id {
			t.Fatalf("%s = %d, %v (want %d)", tc.name, got, err, tc.id)
		}
		state, err := BlockStateByID(tc.id)
		if err != nil {
			t.Fatal(err)
		}
		back, err := BlockStateID(state.Name, state.Properties)
		if err != nil || back != tc.id {
			t.Fatalf("roundtrip %d = %d, %v", tc.id, back, err)
		}
	}
}
