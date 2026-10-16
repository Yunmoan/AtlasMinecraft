# Java 1.21.1 world generation reference

`internal/worldgen` is a pure Go port of selected generation primitives from
the supplied `jar/server.jar`. The JAR is used only to build and verify test
vectors. It is not used when Atlas runs.

Generation version 12 adds stronghold ring starts, all twelve piece types and
normal chunk/Anvil integration. It also adds existing armor-stand equipment
interaction and single-chest menus, and fixes dynamic-light packet ordering.
See [the version-12 validation report](../../artifacts/worldgen-v12-validation.md)
for independent references, final binary checks and remaining coverage boundaries.

Generation version 11 added pillager outposts and desert, plains, savanna, snowy
and taiga villages to normal Overworld generation. This includes terrain
projection, processors, village feature elements, containers and template actors.
Camel placement uses its precise entity position for biome and Anvil ownership,
and the world age for pose finalization. Structure actor gameplay remains incomplete.
See [the version-11 validation report](../../artifacts/worldgen-v11-validation.md).

Generation version 10 added End City starts, templates, loot chests, Shulkers and
Elytra item frames to normal End generation. It also settles generated bubble
columns and leaf distances after decoration. Nether and End cold terrain halos
share a bounded worker budget, while decoration retains its existing order.
See [the version-10 validation report](../../artifacts/worldgen-v10-validation.md).

Version 9 fixed biome-boundary decoration admission: the placement
filter uses the seeded block-position BiomeManager selection before testing
feature membership. Its 18,432 native-JAR boundary cases include river/ocean,
Nether and End transitions. River and ocean noise, surface and carver regression
coverage is documented in [river_ocean_regression.md](river_ocean_regression.md).
Existing saved chunks retain their terrain and generation version.

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
at runtime. `terrain_cases_1.21.1.txt` adds 45 mountain, plateau and cave chunks,
with independent noise, surface and carving JAR vectors in `terrain_*_1.21.1.txt`.
These checks cover the selected seeds/chunks, not every seed or world location.

Paper's archived 1.21.1 branch is pinned at
`3cb8529bdbeeec9ff51df2c94e93dbe7e41bb811` under
`artifacts/worldgen-probes/paper-1.21.1`. The reviewed Moonrise/chunk-system,
bedrock and End patches do not establish a different default Overworld mountain
or plateau density algorithm. The official 1.21.1 JAR remains the terrain oracle;
this is not a complete Paper-world comparison.

`FeatureOrderProbe` exports the Overworld's 151 globally sorted placed features
and each biome's feature list. `OverworldFeatureSchedule` reproduces the
cross-biome, per-step ordering for five JAR-verified biome combinations.
`build_feature_data.py` extracts all 233 placed and 196 configured feature
definitions from the 1.21.1 JAR; Atlas embeds them without a runtime JVM.
`WorldgenRandom` reproduces the Xoroshiro-backed decoration/feature seeds and
legacy-style random calls. The placement probe checks 992 candidate positions
across ten feature types/configurations and six seed/chunk combinations.
`OverworldFeaturePositions` additionally filters by biome membership. The
version-7 runtime compiles all 151 top-level Overworld placed-feature plans;
the [current coverage report](../../artifacts/worldgen-feature-coverage-v12.json) also records
the remaining generation stages and independent beehive random-source difference.
This is plan coverage rather than complete-world equivalence. `PlaceOreFeature` ports the
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
with the JAR for 3,890 registered states. `FaceFullProbe` dumps the per-state
top-face-full collision and non-empty-fluid vectors used by the spawn column
scan; they extend `internal/world/heightmap_flags.txt` alongside the
motion-blocking and ocean-floor vectors. `SpawnSearchProbe` records vanilla's
deterministic climate-based initial spawn search (`Climate$SpawnFinder`) for
nine fixed seeds (including the reported seed), and `FindSpawnPosition` reproduces it exactly for world
creation. `BlueIceProbe` checks 16
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
mushrooms use the original generation-time brightness and survival rules,
verified by `TestMushroomFeatureAgainstVanillaJAR`. Melon patches are enabled
and verified by `LandFeatureProbe`.

The following additional probes compare actual configured/placed features,
including changed block states, success/failure and subsequent random state.
The configured probes use synthetic worlds; their success does not validate
every interaction in a fully decorated neighboring-chunk pipeline.

| Probe | Cases | Behavior |
| --- | ---: | --- |
| `FrozenSpringProbe` | 1,120 | Springs, ice spikes and surface freezing |
| `CaveFeatureProbe` | 10,752 | Dripstone, geodes, lava lakes, lush vegetation, vines and rooted azalea |
| `CavePlacementProbe` | 1,024 | Complete lazy cave placement and selectors |
| `OceanFeatureProbe` | 2,112 | Seagrass, kelp, sea pickles and coral |
| `OceanPlacementProbe` | 2,496 | Ocean placement and coral selectors |
| `IcebergFeatureProbe` | 320 | Packed/blue icebergs and carving |
| `IcebergPlacementProbe` | 320 | Iceberg placement |
| `IcebergRadiusProbe` | 2,820 | Radius computation and floating-point boundaries |
| `LandFeatureProbe` | 1,792 | Bamboo/podzol, melon and block piles |
| `DesertWellProbe` | 256 | Sandstone wells, suspicious sand and loot NBT |
| `MonsterRoomProbe` | 224 | Room validation, protected blocks, chest orientation, loot and spawner NBT |
| `SculkVeinProbe` / `SculkVeinPlacementProbe` | 640 / 128 | Specialized sculk spreader, corner wrapping, fluids and protected supports |
| `PositionSetProbe` | 1,544 | Java HashSet traversal and tree-bin iterator removal |

Sculk veins are enabled in deep-dark runtime generation. The integration test
generates chunk (-256,232), seed 1234, and checks its deep-dark biome and resulting
veins. The initial feature can grow in flowing water; spreading requires the
source-water fluid type. Full sculk-patch worldgen charge propagation is now
implemented and checked against 264 official feature cases. Scattered growths
include catalyst, sensor and shrieker NBT checked against official saved defaults
and actual Anvil reload/replacement. Gameplay sculk ticks remain separate.

Iceberg trig uses Java-compatible fdlibm kernels where Go and Java differ by
one ULP. `feature_trig.go` retains the source license. Java identity ordering
for positions with identical full hashes is still not fully reproduced.
`BlockEntityPacketProbe` independently checks the official chunk-packet entity
header and update tag. Chest (registry ID 1), spawner (9), brushable block
(39), beehive (33), sensor (34), catalyst (36) and shrieker (37) entities are
persisted and sent; this does not implement container menus,
spawner mob ticks or all existing Java block-entity types.

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
python docs/worldgen/build_ocean_tags.py
python docs/worldgen/verify_reference.py
python docs/worldgen/verify_reference.py --probe VegetationProbe --vegetation-set columns
python docs/worldgen/verify_reference.py --probe RuleDiskProbe
python docs/worldgen/verify_reference.py --probe LightEngineProbe
go test ./internal/worldgen -timeout=10m
$env:ATLAS_SERVER_JAR = (Resolve-Path jar/server.jar).Path
go test ./internal/world -run TestNewWorldMetadataOpensInVanillaServer -v -count=1

# Real Atlas binary: default normal creation, protocol batches, edits and restart
./build.ps1 -Output atlas-worldgen.exe
go build -o atlas-worldgen-test.exe ./cmd/atlas-test
$env:ATLAS_SERVER_EXE = (Resolve-Path atlas-worldgen.exe).Path
$env:ATLAS_TEST_EXE = (Resolve-Path atlas-worldgen-test.exe).Path
go test ./cmd/atlas -run 'TestNormalWorldBinaryCreateAndRestart|TestIdleChunksBinaryUnloadAndReload' -v -count=1 -timeout=5m
```

`verify_reference.py --update` deliberately rewrites the checked-in vectors.
The normal verification command checks them without editing files.

Normal-world creation and chunk integration now run `OverworldGenerator`.
It generates a terrain-only 5x5 halo and replays the neighboring 3x3 feature
sources in a fixed order. Ores, forest rocks, blue ice, clay/gravel/sand/grass disks,
supported cave/frozen/ocean/land features, desert wells, monster rooms,
vegetation and the seven fully supported placed tree distributions (`trees_taiga`,
`trees_grove`, `trees_snowy`, `trees_water`, `trees_windswept_forest`,
`trees_windswept_hills`, `trees_badlands`)
are enabled; the requested chunk alone is persisted, including quart biomes,
heightmaps, fluid post-processing positions, generated block entities and
explicit skipped-feature tags.
The immutable base cache coalesces concurrent requests and retains at most 128
completed terrain chunks. Workers compute missing terrain before waiting on
neighbors already in progress, allowing shared halos to use the worker pool.
Cold halos now use a shared terrain worker budget, including the first spawn
scan and canonical structure replay. The runtime budget follows
`chunk-generation-workers`; concurrent final chunks do not multiply that budget.
Preceding structure stages retain at most 128 center-chunk snapshots with private
write overlays. Mineshaft replay prepares only clips containing pieces, preserving
canonical piece flags, spider spawners, loot requests and random consumption.
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
The version-10 output hashes (including generated block entities) for three
adjacent/negative-coordinate chunks are checked in both sequential and parallel
generation. To reproduce the cold
nine-chunk performance measurement (one operation is the entire area):

```powershell
go test ./internal/worldgen -run '^$' -bench '^BenchmarkOverworldArea$' -benchtime=1x -count=2
```

The further performance measurements, current-source baseline and profiles are
recorded in [the October 1 validation report](../../artifacts/chunk-performance-v2/validation.md).
The October 2 structure/halo, template and initial-animal optimizations are
recorded in [the current performance report](../../artifacts/worldgen-probes/performance-v8.md).
The region store now reads compressed records by offset and uses shared read
locks scoped to regions, with codec work outside the file locks. Immutable
network packets have a bounded actor cache; lighting reuses private work arrays
under the engine mutex while every published light snapshot remains detached.
Region rewriting now streams original neighboring records, padding and
timestamps through a fixed copy buffer. Each 30-second autosave unloads saved
chunks idle for one minute outside player views and a one-chunk border, releasing
both packet caches and dense lighting columns. Failed saves retain data; region
actors themselves are retained. See [the version-5 validation report](../../artifacts/worldgen-v5-validation.md)
for controlled heap replay and the real autosave/unload/reconnect binary test.

The opt-in official-server test loads both Atlas-created normal and flat
chunks and checks that a block edit, all quart biomes and 30 generated tree/vegetation
blocks survive actual forced loading and saving. Separate well/monster-room
cases generate official configured features on synthetic terrain, bridge their
entities into Atlas chunks, and verify that the actual server retains 2 well
entities and 3 monster-room entities. The JAR is
used only for tests. New normal metadata enables supported structures and marks
the generator incomplete. Desert pyramids, jungle temples, swamp huts, igloos,
buried treasure, six Overworld ruined-portal variants and both shipwreck variants
join starts, terrain, blocks and canonical saved references in the runtime
pipeline. Nether fortresses, Nether fossils and Nether ruined portals have
dimension generation, references and Anvil reopen checks. End cities include
all 20 templates, rotated/clipped block placement, loot requests, Shulkers and
Elytra item frames. Generated source-water columns above magma/soul sand settle
into bubble columns; generated leaves settle their distance across neighboring
tree bounds. These are generation postprocessing passes, not a general fluid,
leaf-decay or scheduled-tick implementation. The remaining Overworld structure
families still require their own complete starts, processors and pieces.
All 37 Nether and 5 End feature plans and their three-dimension
runtime access are integrated. Generated crystals and supported structure NPCs
are persisted in independent entity Anvil regions and delivered after terrain;
Eighteen initial animal types and supported structure actors are generated;
AI, physics, interactions and spawning beyond that coverage remain incomplete.
Initial spawn uses the JAR-verified climate search and vanilla's
11x11 chunk spiral, with additional Atlas clearance and hazard checks. Generation coverage is version 11;
[the machine-readable report](../../artifacts/worldgen-feature-coverage-v11.json)
lists 151/151, 37/37 and 5/5 feature plans plus integrated/remaining structures,
not a world-equivalence percentage.
Fixed-order decoration
replay has not been proven equivalent to vanilla's neighboring chunk ordering;
the terrain-stage JAR matches are not a claim of identical full worlds.
