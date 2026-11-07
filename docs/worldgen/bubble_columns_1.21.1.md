# Generated bubble columns (Java 1.21.1)

Generation version 15 removes the early bubble-column finalizer. Generated water
stays water until chunk activation and scheduled block execution create bubbles.
Existing saved chunks retain their blocks. The runtime is pure Go; the pinned
official JAR supplies independent test results only.

## Official behavior

The official `MagmaBlock` and `SoulSandBlock` schedule their bubble check after
20 ticks on placement and when their upward neighbor becomes a water block.
Their tick invokes `BubbleColumnBlock.updateColumn` above the support.
The [Minecraft Wiki](https://minecraft.wiki/w/Bubble_Column) also describes the
20-game-tick creation delay and five-game-tick removal delay. The runtime applies
the original scheduled handler, including these column rules:

- Magma produces `bubble_column[drag=true]`.
- Soul sand produces `bubble_column[drag=false]`.
- A bubble support preserves its drag direction when the shared update helper
  is called.
- Only a water block containing source water (`level=0`) or an existing bubble
  can be replaced. Flowing/falling water, kelp, kelp plants, seagrass, tall
  seagrass and waterlogged slabs stop the column.
- Propagation runs vertically upwards. The first unchanged-state write is
  ignored, but an unchanged state further up returns false from the actual
  `Level.setBlock` and stops that invocation.

Bubble shape updates schedule water first, then a block tick after five ticks
if support is invalid, the downward neighbor updates, or an upward update sees
source water rather than a bubble. The delayed handler reads current support;
column writes use flags 2. Magma/soul-sand onPlace schedules after 20 ticks.

WorldGenRegion automatically marks `BlockState.hasPostProcess` positions after
accepted writes. Original registry queries identify magma and the two mushroom
states; soul sand does not have this flag. The Go region retains marks through
overwrites, structure recording/replay and detached copies. Activation runs shape
checks and schedules magma's tick. It does not synthesize a delay-zero generation
tick on magma placement.

LakeFeature separately schedules delay-zero cave-air ticks after upper-cavity
writes, before marking blocks above for postprocessing. ProtoChunkTicks retains
requests even beyond build height when the corresponding write was rejected.
These requests survive generation and Anvil; cave air has an empty tick handler.
Original `WorldGenTickAccess` routes both typed block and fluid requests by X/Z
chunk ownership without feature-write bounds filtering; missing cached chunks
produce an error rather than silently discarding requests.

## Independent verification

`BubbleColumnProbe.java` calls the real 1.21.1 `BubbleColumnBlock.updateColumn`
and magma/soul-sand `updateShape` with actual registered block/fluid states.
Its synthetic world's setBlock contract explicitly follows official
`Level`/`LevelChunk`: writing an unchanged state returns false. It contains no
reimplementation of the bubble algorithm.

The fixture contains 960 column updates and 48 actual neighbor-scheduling
observations. Columns cover water levels 0 through 15, magma, soul sand,
existing upward/downward bubbles, unsupported bases, vegetation, waterlogged
blocks, obstacles, adjacent water, negative coordinates, chunk edges and world
height bounds. All 960 Go helper updates match the JAR's changed-state count
and SHA256. The 48 neighbor rows now verify actual scheduling rather than merely
counting observations. An additional 312 original scheduled-tick/onPlace cases
check writes, flags, order, delay and unchanged RNG, and 144 full-state shape cases
check bubble scheduling. All 26,684 hasPostProcess states match in seven defined
neighbor contexts. LakeFeature matches 512 original cases and 44,550 ordered
requests; the cave suite matches 10,752 direct cases and 1,024 placement cases,
including automatic marks and tree
boundary fluid requests previously omitted by its defined native adapter.

`native_bubble_columns_1.21.1.txt` is independently extracted from the official
server's full save at seed 4511451270791457643: chunk (0,10) contains 15 downward
bubble blocks and chunk (-3,44) contains 3. The actual normal-world test creates
a fresh save, opens it, generates both neighborhoods, advances 100 explicit world
ticks, compares all 18 locations and
directions in both loaded and Anvil states, closes and reopens it, then checks
that a player's stone replacement survives another restart.

The official reference uses disabled structures and the comparison's persisted
decoration salt 5393695769601225306. This coordinate fixture does not establish
complete decorated-world equality or graphical-client rendering.

The live original-server fixture now covers 28 boxes and 17 checkpoints, including
ticks 19, 20, 21 and 25. Bubble scenes cover creation, direction reversal and
support removal. All 1,949,696 states, block-entity and falling-entity NBT,
5,395 fluid records and 245 block records match, including heap order. Both frozen ocean arrays match
all 196,608 states and 133 delay-zero generation block requests (92 cave-air,
41 leaves); both complete arrays still match after 100 world ticks.

Default server-clock/ticket integration, specialized falling entities, other delayed
families, random leaf decay and bubble effects on entities remain incomplete.
The [falling-block runtime](falling_blocks_1.21.1.md) now reproduces all 217 ordered
ocean activation requests, including 77 sand/gravel requests.
Version 16 preserves native ProtoChunk postprocessing append
order and repeated positions through generation, caches and Anvil; see the
[ordered-list reference checks](postprocessing_1.21.1.md). Mixed activation queue
equivalence beyond these samples remains to be established. The generator stays incomplete.

Rebuild and verify:

```powershell
python tools/worldgen/build_bubble_column_probe.py
python tools/worldgen/build_bubble_tick_probe.py --check
python tools/worldgen/build_lake_tick_probe.py --check
python tools/worldgen/build_fluid_runtime_fixture.py --check
go test ./internal/worldgen -run '^Test(Bubble|GenerationPostprocessing|GenerationAutomaticPostprocessing|LakeGeneration)' -count=1
go test ./internal/world -run '^Test(GeneratedChunksAndTicksAgainstFrozenVanillaWorld|GeneratedBubbleColumnsPersistAndPreserveEdits|FinalOceanChunksAgainstVanillaServer|FluidRuntimeBoxesAgainstVanillaServer)$' -count=1
```

Current lifecycle evidence is under `artifacts/generation-block-reference/`.
The original arena publication is `server-1h4oiqp_` under
`artifacts/fluid-runtime-reference/`. The earlier primitive evidence below remains
useful; its old early-generation benchmark is no longer a production measurement.

Evidence logs:

- `artifacts/river-ocean-debug/bubble-column-oracle-build.txt`
- `artifacts/river-ocean-debug/bubble-column-jar-test.txt`
- `artifacts/river-ocean-debug/bubble-column-generation-restart-test.txt`
- `artifacts/river-ocean-debug/bubble-column-test-benchmark.txt`
- `artifacts/river-ocean-debug/bubble-column-race.txt`

SHA256 fixtures:

- Synthetic JAR: `37c955a5a17ac770a1143dd2d7b7e4ea11fac718a5c333c56991ab4bf7f910dd`.
- Full native save: `e0810229cda1e3c0b1c42202d4e683e97de44ef51f6244aa2747263fe995709f`.
