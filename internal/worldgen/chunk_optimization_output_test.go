package worldgen

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

// These whole-result digests are captured from d5b2aa2 before the performance
// changes. They protect block/property/biome arrays, ordered postprocessing and
// tick requests, block entities, structures and authored entity requests. They
// establish output preservation, not independent vanilla compatibility.
func TestChunkOptimizationOutput(t *testing.T) {
	wanted := map[string]string{
		"overworld/seed=1234/structures=false":       "c50b5e982582902d35ccf30617b5f177352317ff6a03f3ad4034b79cf2f51882",
		"overworld/seed=1234/structures=true":        "4c10a1e234b578c1df7ee7e87136e02627f92aa4258df1a9bd3a06186e0826e9",
		"nether/seed=1234/structures=false":          "5ffa4756b03c5cf1315a80106328790fb5d79f4f084283b58cdad664cbfa013f",
		"nether/seed=1234/structures=true":           "b3df21cee36e65e55f99a3060e4d6f86daa014b4cc708212ecad8796a071a5d1",
		"end/seed=1234/structures=false":             "5a015997c0886988ef76d6b761ac64f1ca957ce136b96a14735433bec395f30f",
		"end/seed=1234/structures=true":              "1888ce1f5f97e179acc3b54aa8551321d6b3b3720ffedc399b140e5e373c88aa",
		"overworld/seed=-987654321/structures=false": "c335a42209ab0daf3dc96206cabf0bc32e0eee31b4f39e2f0e3dcca9aeaa2141",
		"overworld/seed=-987654321/structures=true":  "37477e99bb28516941ad4ea89ffee33606551e8603f3612f98b8aea027c412b3",
		"nether/seed=-987654321/structures=false":    "bafe0ffc1ea5f48d54c143b359140c18056d2b833528ec4fad9bd32d183d49d0",
		"nether/seed=-987654321/structures=true":     "d1a2ee32ecf3ffa6ca44c83f1edd734b1b54ce684f5d87ff2ac21cf4baef3664",
		"end/seed=-987654321/structures=false":       "587fca8908a2ff3313bea1f71d3729a5d2661a2aef3e9f0e7a0c01c785da1c3d",
		"end/seed=-987654321/structures=true":        "27b61b8feaa9dcf61b331e61518c37f3f1ecc32449022353c1f98542e4977445",
	}
	for _, seed := range []int64{1234, -987654321} {
		for _, dimension := range []string{"overworld", "nether", "end"} {
			for _, structures := range []bool{false, true} {
				name := fmt.Sprintf("%s/seed=%d/structures=%t", dimension, seed, structures)
				t.Run(name, func(t *testing.T) {
					var result any
					if dimension == "overworld" {
						g, err := NewOverworldGeneratorWithOptions(seed, OverworldGeneratorOptions{GenerateStructures: structures})
						if err != nil {
							t.Fatal(err)
						}
						result, err = g.Generate(40, -40)
						if err != nil {
							t.Fatal(err)
						}
					} else {
						g, err := NewDimensionGeneratorWithOptions(seed, dimension, DimensionGeneratorOptions{GenerateStructures: structures})
						if err != nil {
							t.Fatal(err)
						}
						x, z := int32(-1), int32(0)
						if dimension == "end" {
							x, z = 145, -77
						}
						result, err = g.Generate(x, z)
						if err != nil {
							t.Fatal(err)
						}
					}
					encoded, err := json.Marshal(result)
					if err != nil {
						t.Fatal(err)
					}
					digest := fmt.Sprintf("%x", sha256.Sum256(encoded))
					if digest != wanted[name] {
						t.Fatalf("complete output changed: got %s, want %s", digest, wanted[name])
					}
					t.Logf("%s %s", name, digest)
				})
			}
		}
	}
}
