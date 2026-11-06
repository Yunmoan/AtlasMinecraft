# Structure template updates — Java 1.21.1

Generation version 17 introduced retention of the delayed requests made when a
structure template updates its edge and then its placed blocks. The shared
path now handles sand/gravel and other FallingBlock subclasses, leaves,
magma/soul sand/bubble columns, and plain liquids before delegating other block
shapes to the template's palette-specific adapter. Version 18 also handles all
eight BrushableBlock states and aquatic plant shapes; brushable shape/placement callbacks schedule delay 2.
Their scheduled brushing-reset/falling handler is a separate
[native-checked runtime path](brushable_runtime_1.21.1.md).
See [real ocean-ruin saves](ocean_ruin_runtime_reference_1.21.1.md).

Previously, for example, one sand block at (-1,60,15) produced no generation
requests through the igloo adapter. Original `StructureTemplate.updateShapeAtEdge`
produces six ordered sand requests in that case. The owning ProtoChunkTicks
container keeps the first request for the block type/position and packs delay 0.
These are queued updates, not an early terrain settling pass.

## Original source and independent capture

The reference remains Java 1.21.1, DataVersion 3955, server SHA-1
`59353fb40c36d304f2035d51e7d6e6baa98dc05c`. The native probe invokes
`StructureTemplate` (`ent`).`updateShapeAtEdge` with actual
`BitSetDiscreteVoxelShape` (`exf`) masks, followed by `Block`
(`dfy`).`updateFromNeighbourShapes` when finalization is requested. It records
the original callbacks and feeds them into actual `ProtoChunkTicks` (`eyr`).
No Go update code creates expected states or queue entries.

The authored level supplies raw WorldGenRegion-style writes, an Overworld
dimension and a defined surrounding state. It does not recursively run
ServerLevel placement/neighbor updates. The capture covers:

- 386 registered states: all FallingBlock, BrushableBlock and LeavesBlock states,
  magma/soul sand/bubble columns, all water/lava levels, and every kelp, kelp-plant,
  seagrass and tall-seagrass state.
- 12 surroundings: air, stone, log, leaves, source water, sand, magma, soul
  sand, bubble column, flowing water, flowing lava and waterlogged leaves.
- Single-block, cross-chunk L and hollow 3×3×3 shell masks, with edge-only and
  edge-plus-finalization captures for each.

All **27,792 cases, 2,442,730 ordered requests and 766,346 packed requests** match
through the six template adapters and the shared runtime adapter. Comparisons
include every resulting block, source/flowing fluid identity, callback order,
delay, priority, per-chunk ownership, first-wins deduplication and native packing.
The production structure recorder is also checked for all 386 states in the
cross-chunk L case with both direct and delegated region recording (772 runs).
Each adapter also matches the original world's RNG continuation, including
kelp-body conversion to a randomly aged head. Another 972 direct kelp shape
callbacks match state, request order and RNG continuation in six surroundings.

LiquidBlock's original update checks the FluidState of both blocks. The old
template helper missed sources carried by bubble columns. The shared liquid
callback now uses the exhaustive native fluid-state table; flowing water/lava
can therefore schedule correctly beside bubble columns or waterlogged blocks.
An additional 1,152 direct liquid shape callbacks match the original registry
oracle. Existing 2,824 tree cases still match 402,367 ordered requests and
185,616 packed requests.

## Storage and execution boundary

Generation regions and structure recorders own these queues. They retain raw
callbacks until their existing first-wins containers deduplicate them. The
existing structure cache, chunk snapshots and native Anvil serializers carry
the typed entries into activation. Physical updates execute later through
explicit runtime ticks. A concrete-powder shape callback can still immediately
return concrete, as the original does; the probe checks that final state too.

Template implementations using `knownShape=true`, including the SinglePoolElement
path, continue to preserve authored connections and skip the template shape
passes. This change does not add blanket updates to every generated structure.

The capture tests the original edge/finalization primitives in defined terrain.
It does not establish complete queue equivalence for every fully decorated
structure or arbitrary seed. Other delayed block families, broader template
palette behavior, default server clock/tickets and specialized entity behavior
remain unfinished. Generation version 18 must not be described as exact
whole-world generation.

```powershell
python tools/worldgen/build_template_tick_probe.py --check
go test ./internal/worldgen -run '^Test(TemplateGenerationTicksAgainstVanilla|TreeGenerationTicksAgainstVanilla|LiquidNeighborShapesAgainstVanilla)$' -count=1 -v
```

Fixture: `testdata/worldgen/template_ticks_1.21.1.jsonl.gz`.
Native captures and validation logs: `artifacts/template-tick-reference/`.

## Regression validation

The native template capture was regenerated and checked independently. All 187
basic tests, targeted cache/recording race tests and `go vet` for worldgen/world
passed. Complete Overworld and dimension structure pipeline tests also passed,
including all five village variants, outposts, mansions, ancient cities, trial
chambers, trail ruins, fortresses and End cities.

Full template block oracles also pass: 2,304 igloo cases, 3,088 End-city calls,
7,008 mansion calls, 8,640 shipwreck cases, 10,368 ocean-ruin cases and 22,464
ruined-portal cases. The last three reference files were initially absent;
`verify_structure_reference.py --update --probe <ProbeName>` rebuilt them from
the pinned official JAR before their checks were rerun successfully. Their
separate `*-blocks-tests.txt` logs supersede the missing-file failures in
`structure-tests.txt`; no expected block results were derived from Go output.

The existing server fixtures still match all 28 runtime boxes at 17 checkpoints
and both final ocean chunks after 100 explicit ticks. Frozen-generation checks
match 196,608 block states and 133 delay-zero block ticks; Anvil reload retains
player edits and repeated postprocessing entries. These are bounded regression
fixtures, not evidence for arbitrary structure-bearing world equality.
