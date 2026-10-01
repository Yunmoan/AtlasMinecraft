# Java 1.21.1 world generation reference

`internal/worldgen` is a pure Go port of selected generation primitives from
the supplied `jar/server.jar`. The JAR is used only to build and verify test
vectors. It is not used when Atlas runs.

Mojang's official 1.21.1 server mappings identify these classes in the
bundled server JAR:

| Vanilla class | Bundled class | Go implementation |
| --- | --- | --- |
| `XoroshiroRandomSource` | `dzz` | `random.go` |
| `LegacyRandomSource` | `dyz` | `legacy_random.go` |
| `ImprovedNoise` | `eof` | `improved_noise.go` |
| `PerlinNoise` | `eoi` | `perlin_noise.go` |
| `NormalNoise` | `eoh` | `normal_noise.go` |
| `BlendedNoise` | `eoe` | `blended_noise.go` |

The 60 `worldgen/noise` parameter files are extracted from the same JAR into
`internal/worldgen/noise_parameters.json`. The Java probes under this directory
produce fixed output for multiple seeds, octave settings, positions, and
coordinate ranges. Go tests compare integer and floating point **bits** against
those outputs.

`build_density_data.py` extracts all 35 built-in density functions and all 7
noise settings (including their surface rules) into `density_data.json`.
`LoadDensityFunction` and `LoadDensitySettings` resolve and validate every
built-in density reference, router field, and nested spline. `DensitySampler`
directly samples basic arithmetic, named noise, climate shifts, gradients,
splines, `old_blended_noise`, and cave `weird_scaled_sampler` nodes. The JAR
probes compare 84 climate, 36 blended terrain, 144 cave sampler, and 9 spline
values bit for bit with Go. Direct sampling rejects nodes
that require `NoiseChunk` cell interpolation or other generation context.

`BiomePresetProbe.java` exports the 7,593 built-in overworld climate parameter
points from the same JAR. `SelectOverworldBiome` reproduces the vanilla
six-parameter nearest-point choice for the probe's independent vectors, and
`OverworldBiomeSampler` combines that table with the Go density sampler at
quart positions. This is a biome-selection stage, not a finished chunk
generator.
For an additional check against actual generated biomes, set
`ATLAS_VANILLA_BIOME_REFERENCE_WORLD` to a stable copy of a default 1.21.1
noise-world save and run `go test ./internal/world -run TestOverworldBiomeSamplesAgainstVanillaChunk`.
The test compares 64 quart positions in chunk (0,0) at four heights.

`NewCellDensitySampler` evaluates `interpolated` nodes on each noise setting's
cell grid (4 by 8 blocks in the default Overworld). `GenerateNoise` implements
the complete Overworld noise fill, including density interpolation, aquifers,
ore veins, and fluid update flags. The JAR probe compares 12 full chunks.
`ApplySurface` applies the built-in surface rule tree, badlands pillars, and
frozen-ocean icebergs. The JAR probe compares 96 fixed-biome and 12 natural-biome
full chunks. `ApplyCarvers` follows the JAR's 17 by 17 source-chunk traversal,
three Overworld configured carvers, legacy random streams, masks, aquifer
replacement, and surface repair. `GenerateCarved` combines those three stages;
the JAR probe compares 12 natural-biome carved chunks, including per-carver
mask counts and all block states. The probes load the built-in replaceable
block tag from the JAR; Atlas embeds its 1.21.1 members and never reads the JAR
at runtime. These checks cover the selected seeds/chunks, not every seed or
world location.

`FeatureOrderProbe` exports the Overworld's 151 globally sorted placed features
and each biome's feature list. `OverworldFeatureSchedule` reproduces the
cross-biome, per-step ordering for five JAR-verified biome combinations.
`build_feature_data.py` extracts all 233 placed and 196 configured feature
definitions from the 1.21.1 JAR; Atlas embeds them without a runtime JVM.
`WorldgenRandom` reproduces the Xoroshiro-backed decoration/feature seeds and
legacy-style random calls. The placement probe checks 992 candidate positions
across ten feature types/configurations and six seed/chunk combinations.
`OverworldFeaturePositions` additionally filters by biome membership, but
most configured feature types are not implemented. `PlaceOreFeature` ports the
Overworld ore algorithm and its air-exposure rule; JAR probes compare 12
synthetic stone-world veins block for block and 32 replacement decisions.
`OverworldFeatureRegion` applies feature edits across already-generated neighboring
chunks. The JAR also validates 3,072 OCEAN_FLOOR_WG column heights from 12
carved chunks used by ore placement. `ForestRockProbe` checks 8 heightmap-based
candidate groups and 20 synthetic-world placements, including failed
placement and cross-chunk coordinates. `PlaceForestRock` uses the JAR's
base-stone/dirt tags and preserves its random state. `MotionBlockingProbe`
classifies all 26,684 block states for MOTION_BLOCKING and OCEAN_FLOOR_WG,
including waterlogged variants, and checks 3,072 carved-chunk motion-blocking
column heights. `build_motion_blocking.py` reduces those JAR vectors to pure-Go
classifiers; the feature region computes and invalidates both heights as blocks
change. The previous non-air/non-fluid ocean-floor approximation disagreed
with the JAR for 3,890 registered states. `BlueIceProbe` checks 16
synthetic-world cases block for block and verifies subsequent random state.
The region rejects a missing 3x3 neighborhood before applying ores, forest
rocks, or blue ice. This is not yet a complete decoration pass.

`FeatureStreamProbe` checks the lazy placement pipeline with 60 independent
JAR cases, consuming variable amounts of random state between positions.
Runtime decoration uses this pipeline; the older candidate-list probes check
isolated modifiers and do not validate full feature execution order.

`TreeFeatureProbe` checks 600 configured tree cases against the actual JAR,
including ordinary/fancy oak, birch, spruce and pine, obstructions, existing
logs/leaves, source/flowing water, negative/cross-chunk positions and extreme
coordinates/heights. It binds the JAR's block tags before testing. Full block
state hashes include log axes, leaf distances and waterlogging, together with
the result and subsequent random state. Leaf-distance updates reproduce Java
HashSet bucket traversal and vanilla's repeated pending-distance writes.
`TreePlacementProbe` checks 1,008 complete lazy placement/selector cases over
flat/sloping grass, water, stone and snow/powder-snow terrain, including zero
counts, failed nested survival checks and live heightmap changes between trees.
Nested pine/spruce snow placements also test successful upward scans through
powder snow and rejection when the search exceeds the eight-block limit.
Both probes use the actual built-in feature definitions; the Go runtime is
still standard-library-only and never starts Java.

`VegetationProbe` checks 4,208 configured and top-level vegetation cases,
including weighted, noise, threshold-noise and dual-noise state providers,
six-draw patch offsets, live heightmaps, substrate checks and double plants.
It covers flat/sloping grass, sand, podzol, stone, water, blocked upper halves,
negative/cross-chunk positions and build-height/extreme-coordinate boundaries.
`vegetation_features_1.21.1.txt` lists the original 25 enabled top-level features
and their step-9 indices. `column_vegetation_features_1.21.1.txt` adds cactus,
sugar cane, waterlilies, huge mushrooms and mushroom-island selection. These
include contextual survival and block-column height/truncation rules. Small
mushroom, melon and other unsupported graphs still reject their branches.

`RuleDiskProbe` compares sand/grass rule-provider disks, including all changed
states, following random state and postprocessing lists, across 96 cases.
`LightRulesProbe` exports emission, opacity and exact face-union occlusion for
26,684 states. `LightEngineProbe` runs the actual Java sky/block engines on
four 3x3 scenarios; Go checks every central-cell value through layer hashes.

`BlockRulesProbe` independently extracts 26,684 state rules and 320 deduplicated
collision shapes for the gameplay foundation into `internal/blockrules`.
These are origin/empty-context shapes; entity- and position-dependent behavior
is not validated by this table. `verify_reference.py --probe BlockRulesProbe`
compares the decompressed embedded data directly with freshly produced JAR output.

To verify with JDK 21 and Python 3:

```powershell
python docs/worldgen/build_noise_parameters.py
python docs/worldgen/build_density_data.py --check
python docs/worldgen/build_feature_data.py --check
python docs/worldgen/build_motion_blocking.py --check
python docs/worldgen/build_tree_tags.py
python docs/worldgen/build_vegetation_tags.py
python docs/worldgen/verify_reference.py
python docs/worldgen/verify_reference.py --probe VegetationProbe --vegetation-set columns
python docs/worldgen/verify_reference.py --probe RuleDiskProbe
python docs/worldgen/verify_reference.py --probe LightEngineProbe
go test ./internal/worldgen -timeout=10m
$env:ATLAS_SERVER_JAR = (Resolve-Path vanilla/server.jar).Path
go test ./internal/world -run TestNewWorldMetadataOpensInVanillaServer -v -count=1

# Real Atlas binary: default normal creation, protocol batches, edits and restart
go build -o atlas.exe ./cmd/atlas
go build -o atlas-test.exe ./cmd/atlas-test
$env:ATLAS_SERVER_EXE = (Resolve-Path atlas.exe).Path
$env:ATLAS_TEST_EXE = (Resolve-Path atlas-test.exe).Path
go test ./cmd/atlas -run TestNormalWorldBinaryCreateAndRestart -v -count=1 -timeout=5m
```

`verify_reference.py --update` deliberately rewrites the checked-in vectors.
The normal verification command checks them without editing files.

Normal-world creation and chunk integration now run `OverworldGenerator`.
It generates a terrain-only 5x5 halo and replays the neighboring 3x3 feature
sources in a fixed order. Ores, forest rocks, blue ice, clay/gravel/sand/grass disks,
supported vegetation and the seven fully supported placed tree distributions (`trees_taiga`,
`trees_grove`, `trees_snowy`, `trees_water`, `trees_windswept_forest`,
`trees_windswept_hills`, `trees_badlands`)
are enabled; the requested chunk alone is persisted, including quart biomes,
heightmaps, fluid post-processing positions and explicit skipped-feature tags.
The immutable base cache coalesces concurrent requests and retains at most 128
completed terrain chunks. Workers compute missing terrain before waiting on
neighbors already in progress, allowing shared halos to use the worker pool.
Tests cover negative coordinates, concurrent repeat
generation, exploration order, block edits, preload and restart.

Runtime startup now prepares a configurable square before opening the listener:
`spawn-preload-radius=3` means 49 chunks; `chunk-generation-workers=0` chooses
1..8 workers from the available Go CPU threads. `-spawn-radius` and `-gen`
override the properties. Existing chunks are loaded unchanged. Preparation
uses bounded concurrent actor requests, reports progress, and can be canceled;
world shutdown waits for in-flight generation and persistence.

Biome lookup uses an immutable interval index, preserving exhaustive-search
distance/tie results (including the sampler's previous-point preference).
Noise interpolation reuses the eight corners within a cell, with unchanged
floating-point evaluation order. Each chunk has its own mutable sampler;
decoration uses sparse private block/postprocessing overlays and fully detaches
the returned chunk. Density operands and feature placement/predicate plans are
compiled once, preserving evaluation order and the lazy random stream. Aquifer
centers use a compact chunk-local grid with a general fallback for probes outside
that grid. Internal surface-rule states are immutable; public sampling results
and final generated chunks receive their own property maps.
The version-4 output hashes for three adjacent/negative-coordinate chunks are
checked in both sequential and parallel generation. To reproduce the cold
nine-chunk performance measurement (one operation is the entire area):

```powershell
go test ./internal/worldgen -run '^$' -bench '^BenchmarkOverworldArea$' -benchtime=1x -count=2
```

The further performance measurements, current-source baseline and profiles are
recorded in [the October 1 validation report](../../artifacts/chunk-performance-v2/validation.md).
The region store now reads compressed records by offset and uses shared read
locks scoped to regions, with codec work outside the file locks. Immutable
network packets have a bounded actor cache; lighting reuses private work arrays
under the engine mutex while every published light snapshot remains detached.

The opt-in official-server test loads both Atlas-created normal and flat
chunks and checks that a block edit, all quart biomes and 30 generated tree/vegetation
blocks survive actual forced loading and saving. The JAR is
used only for tests. New normal metadata disables structures and marks the
generator incomplete. Structure starts/references, decorated/special trees,
remaining features (including small mushrooms), Nether/End
and vanilla spawn search still need ports. Generation coverage is version 4.
Fixed-order decoration
replay has not been proven equivalent to vanilla's neighboring chunk ordering;
the terrain-stage JAR matches are not a claim of identical full worlds.
