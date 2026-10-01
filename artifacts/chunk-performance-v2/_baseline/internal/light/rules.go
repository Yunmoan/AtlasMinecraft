// Package light computes state-dependent sky and emitted block light. Block
// properties and face-union occlusion come from the actual Java 1.21.1 JAR.
package light

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed rules_1.21.1.json.gz
var ruleData []byte

type Rule struct {
	Emission, Opacity uint8
	Faces             [6]uint16
}

type ruleTable struct {
	states  []Rule
	blocked [][]bool
}

var rulesOnce sync.Once
var rules ruleTable
var rulesError error

func loadRules() (*ruleTable, error) {
	rulesOnce.Do(func() {
		reader, err := gzip.NewReader(bytes.NewReader(ruleData))
		if err != nil {
			rulesError = err
			return
		}
		defer reader.Close()
		var raw struct {
			States  [][8]uint16 `json:"states"`
			Blocked [][]bool    `json:"blocked"`
		}
		if err := json.NewDecoder(reader).Decode(&raw); err != nil {
			rulesError = err
			return
		}
		if len(raw.States) != 26684 || len(raw.Blocked) == 0 {
			rulesError = fmt.Errorf("incomplete 1.21.1 light rules")
			return
		}
		for _, row := range raw.Blocked {
			if len(row) != len(raw.Blocked) {
				rulesError = fmt.Errorf("invalid light occlusion matrix")
				return
			}
		}
		for _, row := range raw.States {
			if row[0] > 15 || row[1] > 15 {
				rulesError = fmt.Errorf("invalid state light level")
				return
			}
			rule := Rule{Emission: uint8(row[0]), Opacity: uint8(row[1])}
			for i := range rule.Faces {
				if int(row[i+2]) >= len(raw.Blocked) {
					rulesError = fmt.Errorf("invalid light face index")
					return
				}
				rule.Faces[i] = row[i+2]
			}
			rules.states = append(rules.states, rule)
		}
		rules.blocked = raw.Blocked
	})
	return &rules, rulesError
}
