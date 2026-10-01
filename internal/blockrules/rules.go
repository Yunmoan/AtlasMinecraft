// Package blockrules contains test-extracted 1.21.1 block-state rules. The
// server uses only Go and embedded data; the JAR is a development oracle.
package blockrules

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

// Box stores min X/Y/Z followed by max X/Y/Z, relative to a block position.
type Box [6]float64

//go:embed rules_1.21.1.json.gz
var ruleData []byte

var once sync.Once
var data struct {
	Shapes [][]Box     `json:"shapes"`
	States [][2]uint16 `json:"states"`
}
var dataError error

type Rules struct {
	Boxes                                                           []Box
	Replaceable, FallThrough, TopSturdy, Offset, Solid, SolidRender bool
	FaceSturdy                                                      [6]bool // down, up, north, south, west, east
	TopCenter, BottomCenter, ConnectionException                    bool
}

func Lookup(state uint32) (Rules, error) {
	once.Do(func() {
		gz, err := gzip.NewReader(bytes.NewReader(ruleData))
		if err != nil {
			dataError = err
			return
		}
		defer gz.Close()
		dataError = json.NewDecoder(gz).Decode(&data)
		if dataError == nil && (len(data.States) != 26684 || len(data.Shapes) != 320) {
			dataError = fmt.Errorf("incomplete 1.21.1 block rules")
		}
		if dataError == nil {
			for _, state := range data.States {
				if int(state[0]) >= len(data.Shapes) {
					dataError = fmt.Errorf("invalid collision shape index")
					break
				}
			}
		}
	})
	if dataError != nil {
		return Rules{}, dataError
	}
	if state >= uint32(len(data.States)) {
		return Rules{}, fmt.Errorf("unknown block state %d", state)
	}
	entry := data.States[state]
	flags := entry[1]
	rules := Rules{Boxes: data.Shapes[entry[0]], Replaceable: flags&1 != 0, FallThrough: flags&2 != 0, TopSturdy: flags&4 != 0, Offset: flags&8 != 0, Solid: flags&16 != 0, SolidRender: flags&32 != 0, TopCenter: flags&(1<<12) != 0, BottomCenter: flags&(1<<13) != 0, ConnectionException: flags&(1<<14) != 0}
	for face := range rules.FaceSturdy {
		rules.FaceSturdy[face] = flags&(1<<(6+face)) != 0
	}
	return rules, nil
}
