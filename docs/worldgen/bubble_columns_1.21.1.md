# Generated bubble columns (Java 1.21.1)

The final Overworld and dimension generation paths call
`PostProcessGeneratedBubbleColumns` after detaching the requested chunk's
decorated terrain and before the runtime converts it to protocol/Anvil
palettes. Existing saved chunks are loaded unchanged.

## Official behavior

The official `MagmaBlock` and `SoulSandBlock` schedule their bubble check after
20 ticks on placement and when their upward neighbor becomes a water block.
Their tick invokes `BubbleColumnBlock.updateColumn` above the support.
The generation finalizer applies this check's block-state result directly:

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

The pure-Go helper does not implement a general tick scheduler, fluid flow,
block placement timing, entity drag/buoyancy, or subsequent player-created
bubble-column behavior. The generation coverage still records
`generation_lighting_and_ticks` as incomplete.

Only sparse magma/soul-sand source states are visited. A source placed across
a chunk border by a neighboring decoration replay belongs to the destination's
final map. Bubble updates need no horizontal neighbor reads and mutate only
the detached final chunk, preserving cached terrain and independent overlays.
Every resulting state's property map is independent.

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
and SHA256. The generation entry point separately matches 384 magma/soul-sand
cases and is idempotent for those inputs.

`native_bubble_columns_1.21.1.txt` is independently extracted from the official
server's full save at seed 4511451270791457643: chunk (0,10) contains 15 downward
bubble blocks and chunk (-3,44) contains 3. The actual normal-world test creates
a fresh save, opens it, generates both chunks, compares all 18 locations and
directions in both loaded and Anvil states, closes and reopens it, then checks
that a player's stone replacement survives another restart.

The official reference uses disabled structures and the comparison's persisted
decoration salt 5393695769601225306. This coordinate fixture does not establish
complete decorated-world equality or graphical-client rendering.

Rebuild and verify:

```powershell
python docs/worldgen/build_bubble_column_probe.py
python docs/worldgen/build_native_bubble_column_fixture.py artifacts/river-ocean-debug/native-final-server/world
& $env:TEMP/atlas-go/go/bin/go.exe test ./internal/worldgen ./internal/world -run '^(TestBubbleColumnsAgainstVanillaJAR|TestGeneratedBubbleColumnsCrossChunkOverlayIsolation|TestGeneratedBubbleColumnsPersistAndPreserveEdits)$' -count=1
```

Focused race checks passed in both packages. The microbenchmark including
resetting and processing a 42-block column measured 8.043 microseconds/op,
14,784 bytes/op and 88 allocations/op on Windows/amd64, Ryzen 9 8945HX; the
production pass does not allocate a full-volume working array.

Evidence logs:

- `artifacts/river-ocean-debug/bubble-column-oracle-build.txt`
- `artifacts/river-ocean-debug/bubble-column-jar-test.txt`
- `artifacts/river-ocean-debug/bubble-column-generation-restart-test.txt`
- `artifacts/river-ocean-debug/bubble-column-test-benchmark.txt`
- `artifacts/river-ocean-debug/bubble-column-race.txt`

SHA256 fixtures:

- Synthetic JAR: `37c955a5a17ac770a1143dd2d7b7e4ea11fac718a5c333c56991ab4bf7f910dd`.
- Full native save: `e0810229cda1e3c0b1c42202d4e683e97de44ef51f6244aa2747263fe995709f`.
