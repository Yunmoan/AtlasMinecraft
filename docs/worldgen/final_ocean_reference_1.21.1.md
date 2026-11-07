# Final ocean chunks from the official server

`build_final_ocean_fixture.py` creates an isolated save with the pinned official
Java 1.21.1 bundle, seed `4511451270791457643`, structures disabled, and a
localhost-only listener on a temporary port. The existing native bubble-column
coordinates identify the two target chunks: `(0,10)` and `(-3,44)`.

The runner disables random ticks, mob spawning, daylight/weather progression
and spawn chunks before requesting the targets. It freezes ticks, uses
`execute if loaded` to wait for each asynchronous `forceload`, saves a pre-tick
snapshot, steps exactly 100 game ticks, saves and stops. The Go extraction tool
reads only original Anvil palettes; it does not generate expected terrain.
Each row contains all 98,304 block-state IDs, encoded as big-endian uint16,
gzip and base64. The run directory retains console output, command history,
server properties, original `.mca` files and their SHA-256 provenance.

Command behavior is documented by the Minecraft Wiki's
[`forceload`](https://minecraft.wiki/w/Commands/forceload) and
[`tick`](https://minecraft.wiki/w/Commands/tick) pages. Syntax and response text
are also verified against the pinned JAR's `en_us.json`, since current wiki
pages include changes after 1.21.1.

```powershell
python tools/worldgen/build_final_ocean_fixture.py --check
python tools/worldgen/build_final_ocean_fixture.py --check --trace-decoration
go test ./internal/world -run '^TestFinalOceanChunksAgainstVanillaServer$' -count=1
```

The optional Java agent records original `ChunkGenerator.applyBiomeDecoration`
entry/exit and original placed-feature RNG states, plus ore-stage snapshots for
the failing `(-3,44)` case. It does not replace either algorithm. The helper uses
the pinned, SHA-256-checked ASM 9.8 dependency from Maven Central (see the
[ASM project](https://asm.ow2.io/)); generated agent classes and JARs remain in
ignored artifacts. Both uninstrumented and instrumented fresh runs have produced
identical final target snapshots.

## Current diagnostic result

The recovered reference is deliberately an exact comparison, not a tolerated
difference count. Both `(0,10)` and `(-3,44)` now match all 98,304 final states
after 100 explicit world ticks. The initial 168 differences included 154
ore/rock replacements and four short-grass blocks; preserving WG heightmaps
through FEATURES removed those 158 generation differences. Connecting the
fluid queue, activation and live block callbacks removed the remaining ten
later lava-flow differences. The test now belongs to `internal/world`, since
`Generate` alone does not execute 100 ticks. The original fixture is unchanged.

Before the fix, following the recorded original 7x7 decoration calls left the
same 168 differences. An arbitrary order happened to reduce the count, but was
not the correct fix. All 513 starting RNG states matched; only two
`ore_coal_lower` endings differed. Stage snapshots located the first divergence
at `ore_dirt`: its input target chunk was identical, but its output missed 82
dirt blocks. Its first six origins matched; a heightmap-dependent early exit
changed the seventh origin, despite the overall feature ending with the same
RNG state. This is why checking only a feature's final RNG is insufficient.

Original `ChunkStatus.CARVERS` switches `ProtoChunk.setBlockState` to maintaining
only the non-WG heightmaps. `ChunkStatusTasks.generateFeatures` primes those maps;
it does not replace the retained WG maps with current feature-modified heights.
`BeginFeatureStage` now preserves the two WG maps, while the live maps still
follow writes. Borrowed terrain is read lazily; mutable test worlds capture a
column before its first feature write. `OreFeature` explicitly requests
`OCEAN_FLOOR_WG`. The independent `GenerationHeightmapProbe` checks 56 rows using
actual ProtoChunk writes, status transition and heightmap queries; Go also tests
cold reads after all writes and both borrowed/mutable terrain.
After the fix, all 1,026 starting/ending random-state checkpoints and all 198
ore-stage block snapshots match the instrumented original server. The full-box
and optimized mesa-mineshaft replays also remain equal; their Go snapshot hash
changes because cutting exposed corridors no longer lowers `OCEAN_FLOOR_WG`.

The before/after-tick snapshots also distinguish generated terrain from later
simulation: 100 ticks change 15 blocks in `(0,10)` and 41 in `(-3,44)`, including
bubble columns, leaves and the ten lava-flow blocks. Version 14 removes early
leaf finalization; version 15 removes early bubble-column finalization and adds
lake cave-air requests. The frozen-world gate checks both complete arrays and
all 133 delay-zero generation block requests (92 cave-air and 41 leaves).
`TickScheduled` then advances leaves, bubbles and fluids in the final-world gate.

## Generation fluid queues

`GenerationFluidTickProbe` exercises the original `ProtoChunkTicks`, including
30 schedule requests, duplicate type/position keys, priorities and two save
times. Its 20 saved rows show that generation queues keep the first request,
retain insertion order and priority, and save delay zero regardless of the
requested trigger time. Source and flowing fluids have distinct registry keys.
An exhaustive 26,684-state vector also checks `BlockState.getFluidState`,
including waterlogged blocks, aquatic plants and bubble columns.

`PendingFluidTick` now survives feature generation, structure recording and
clipped replay, preceding-stage cache copies, detached chunk copies and Anvil
`fluid_ticks`. It is independent of `PostProcessing`; replacing a block does
not rewrite the type of an already scheduled request. Full-box/optimized mine
digests now include the ordered fluid queue, so those Go baselines change even
when their blocks are unchanged. The older feature oracle's combined update
position count remains checked as a union of both mechanisms.

The runner also preserves the frozen pre-tick world and extracts
`native_generation_fluid_ticks_1.21.1.txt`. It contains a delay-zero lava request
at `(-39,-53,712)` and two delay-five water requests in `(0,10)`. A fresh official
run reproduced all three records. The lava request matches the generated Go
queue and survives Anvil save/reload and block edits. The two water requests
belong to chunk activation, not `ProtoChunkTicks`: original
`LevelChunk.postProcessGeneration` ticks nonempty fluids at marked positions
immediately, then clears its PostProcessing lists. `World.TickFluids` now runs
that activation sequence and the runtime scheduler through live liquid block
callbacks. This resolves the ten final lava-flow differences in the end-to-end
test. The [fluid execution guide](fluid_simulation_1.21.1.md) records the
independent step, collision, callback and scheduler checks and the remaining
limits. The default server clock still needs ticking-ticket policy and broader
block-update coverage before it can enable this path.

The Minecraft Wiki documents the
[`fluid_ticks` record fields](https://minecraft.wiki/w/Chunk_format) and the
distinction between [scheduled fluid ticks](https://minecraft.wiki/w/Tick) and
random ticks. Version-sensitive lifecycle details above are checked against
the pinned 1.21.1 `ProtoChunkTicks` and `LevelChunk` bytecode rather than the
wiki's unverified conversion note.

```powershell
python tools/worldgen/build_generation_fluid_tick_probe.py --check
go test ./internal/worldgen -run '^TestGenerationFluidTicks' -count=1
go test ./internal/world -run '^TestGenerationStageFluidTicksPersistAndPreserveEdits$' -count=1
```

The runner additionally publishes `frozen_ocean_chunk_1.21.1.txt` and
`native_generation_block_ticks_1.21.1.txt` from its pre-tick save; `--check`
independently regenerates and compares these alongside final arrays and fluid
queues. The block fixture deliberately retains all original requests, including
currently unsupported families. The generation gate compares all states and
delay-zero block entries, including persistence after edits. The subsequent
[falling-block runtime](falling_blocks_1.21.1.md) also compares the complete
217-entry mixed activation queue, including exact native heap order. The additional
original records are seven magma updates after 20 ticks and 77 sand/gravel
updates after two ticks. Magma execution and ordered PostProcessing storage are
implemented, as are sand/gravel activation and ordinary falling-entity execution.
Mixed queue equivalence outside these two samples and specialized entity behavior
remain incomplete. The [postprocessing guide](postprocessing_1.21.1.md) distinguishes
primitive/stage list checks from full native activation checks. The repeated run at
`artifacts/final-ocean-reference/server-haq1ckue` reproduced all four fixtures.
