package blockstate

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

//go:embed states.json.gz
var blockStateData []byte

type BlockState struct {
	Name       string            `json:"name"`
	Properties map[string]string `json:"properties"`
	Default    bool              `json:"default"`
}

var stateOnce sync.Once
var stateTable []BlockState
var stateIDs map[string]uint32
var defaultStateIDs map[string]uint32
var stateError error

func loadStates() error {
	stateOnce.Do(func() {
		gz, err := gzip.NewReader(bytes.NewReader(blockStateData))
		if err != nil {
			stateError = err
			return
		}
		b, err := io.ReadAll(gz)
		gz.Close()
		if err != nil {
			stateError = err
			return
		}
		if err := json.Unmarshal(b, &stateTable); err != nil {
			stateError = err
			return
		}
		if len(stateTable) != 26684 {
			stateError = fmt.Errorf("block state table has %d entries", len(stateTable))
			return
		}
		stateIDs = make(map[string]uint32, len(stateTable))
		defaultStateIDs = make(map[string]uint32)
		for i, state := range stateTable {
			stateIDs[stateKey(state.Name, state.Properties)] = uint32(i)
			if state.Default {
				defaultStateIDs[state.Name] = uint32(i)
			}
		}
	})
	return stateError
}

func stateKey(name string, properties map[string]string) string {
	if !strings.Contains(name, ":") {
		name = "minecraft:" + name
	}
	if len(properties) == 0 {
		return name
	}
	keys := make([]string, 0, len(properties))
	for k := range properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	b.WriteString(name)
	b.WriteByte('[')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(properties[k])
	}
	b.WriteByte(']')
	return b.String()
}

func BlockStateID(name string, properties map[string]string) (uint32, error) {
	if err := loadStates(); err != nil {
		return 0, err
	}
	id, ok := stateIDs[stateKey(name, properties)]
	if !ok {
		if baseID, found := defaultStateIDs[stateKey(name, nil)]; found {
			merged := make(map[string]string, len(stateTable[baseID].Properties))
			for key, value := range stateTable[baseID].Properties {
				merged[key] = value
			}
			for key, value := range properties {
				merged[key] = value
			}
			id, ok = stateIDs[stateKey(name, merged)]
		}
	}
	if !ok {
		return 0, fmt.Errorf("unknown 1.21.1 block state %s", stateKey(name, properties))
	}
	return id, nil
}

func BlockStateByID(id uint32) (BlockState, error) {
	if err := loadStates(); err != nil {
		return BlockState{}, err
	}
	if id >= uint32(len(stateTable)) {
		return BlockState{}, fmt.Errorf("invalid block state ID %d", id)
	}
	state := stateTable[id]
	properties := make(map[string]string, len(state.Properties))
	for k, v := range state.Properties {
		properties[k] = v
	}
	state.Properties = properties
	return state, nil
}
