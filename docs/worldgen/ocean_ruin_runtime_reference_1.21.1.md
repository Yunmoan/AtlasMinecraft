# Ocean-ruin generation and activation — Java 1.21.1

Generation version 18 fixes omitted BrushableBlock requests and places ocean
ruins against the live decoration region. It also writes `keepPacked=false`
when converting generated block entities to FULL Anvil chunks, without mutating
cached generation NBT.

## Independent server reference

`build_final_ocean_fixture.py --preset ocean-ruins` launches the pinned official
Java 1.21.1 server (DataVersion 3955, SHA-1
`59353fb40c36d304f2035d51e7d6e6baa98dc05c`) on loopback. Seed 1234 has a cold
ruin at chunk (-30,-56) and a warm ruin at (62,-60). Structures are enabled;
random ticks, mob spawning, weather and daylight progression are disabled.
The game is frozen before either remote chunk is requested. Each asynchronous
forceload is followed by an explicit loaded check and `save-all flush`.

The frozen fixture contains:

| Chunk | Block states | Block entities | Block ticks | Fluid ticks |
| --- | ---: | ---: | ---: | ---: |
| -30,-56 | 98,304 | 18 | 602 | 415 |
| 62,-60 | 98,304 | 6 | 306 | 399 |

Block entities include all 20 archaeology records and four chests, retaining
their original loot-table names, signed 64-bit seeds and NBT types. Only compound
key order and the outer block-entity position index are canonicalized; nested
list and array order is retained. Queue comparisons use the complete ordered
lists, including requests for different block types at the same position.

The world-boundary test loads the full read neighborhood, activates the ticking
ring through production postprocessing, saves, and reads Anvil before comparison.
A frozen FULL save is already postprocessed; comparing it directly to a raw
feature-stage chunk would incorrectly flag a removed upper seagrass half.
The native forceload center activates before the surrounding ring. The replay
loads the center first too; activating a neighbor first can change the heap
layout of otherwise identical fluid requests. No queue sorting hides this
difference.

## Evidence that changed the implementation

The first comparison found 15 missing suspicious-gravel and five missing
suspicious-sand requests. `BrushableBlock` (`dgd`) does not inherit FallingBlock;
its placement and shape handlers independently request delay 2. The expanded
native template oracle covers all eight DUSTED states without changing their
state or executing the delayed handler early.

A second difference was a magma request at (-479,48,-896). The original
`MagmaBlock.updateShape` saw tall seagrass above the block during WorldGenRegion
placement, then air, and only saw water during ServerLevel postprocessing. The
canonical structure cache saw water too early and incorrectly packed a delay-zero
request. Its 596 pending entries differed from the native 595, and although
activation eventually yielded the same 602 request count, the heap save order
was different.

The fix runs each ocean-ruin clip in the current decoration region, using its
original per-clip random streams and private piece state. Neighboring chunks'
earlier decoration can therefore affect callbacks. The frozen block arrays and
loot records remain unchanged while the ordered block queues now match.
Other cached structure families still require equivalent real-save audits.

Adding complete fluid-list comparison exposed omitted aquatic callbacks:
kelp/kelp-plant water requests and seagrass requests, including a position later
overwritten by suspicious gravel. The shared template boundary now uses the
existing native-checked plant callbacks, and the new kelp handler follows
GrowingPlantHeadBlock/BodyBlock (`djn`/`djm`) plus KelpBlock's magma support veto.
An unsupported kelp requests delay 1; head/body conversion can return before the
water request, and a new head draws its age from 0 through 24 using the world's
random source. The 972 direct cases and expanded 27,792 template cases compare
the RNG continuation too. The [wiki's kelp description](https://minecraft.wiki/w/Kelp)
corroborates the randomized age; ordering comes from the fixed-version classes.

Original `LevelChunk.getBlockEntityNbtForSaving` (`dvi.a`) adds the unpacked flag;
the new FULL conversion preserves it. Default block-entity payload probes remain
separate from this chunk-saving wrapper.

## Live scheduler reference

The optional agent observes `ChunkMap.a(ChunkAccess)` before its `isUnsaved`
guard, serializes both real `LevelChunkTicks` through their original save method,
and records `ServerLevel.tickBlock`/`tickFluid` entry before stale-type checks.
It does not replace callbacks, reorder queues, or alter ticket eligibility.
The fixture preserves the native NBT list order and every dispatch's elapsed
tick, phase, position and registered type.

This resolves a misleading disk boundary: the warm chunk's final region file
still contains its initial 306 block and 399 fluid requests, but the live server
executed all 705 and has empty queues. Unchanged chunks can skip saving even
during `save-all flush`. The cold chunk executes 875 block and 935 fluid
requests (1,810 total); its final live queues are empty too.

Ordered dispatch comparison then exposed a terrain-correct discrepancy at
(-477,50,-889): three water requests at elapsed tick 6 had different order.
Native enqueue stacks show `NeighborUpdater.executeShapeUpdate` calling
`Block.updateOrDestroy` (`dfy.a`), then `Level.destroyBlock` (`dcw.a`). An air
shape result takes the destruction path, restoring the previous fluid's legacy
block with flags 3, regardless of the incoming shape flags. Atlas previously
kept flags 2, omitting neighbor notification and reaching the upper water only
in the later shape pass. The runtime now preserves the original destruction
write flags and recursion limit; the six-direction orders are unchanged.
General shape-destruction loot/events remain a separate incomplete boundary.

The published live fixture is
`testdata/worldgen/native_ocean_ruin_live_ticks_1.21.1.json` (SHA-256
`bf3f872a3cd692787bd0fdb2be7bf83840d1929ab95e9410874d8fdf534a643e`).
Source run `server-ecgme7oi` and independent check `server-3ay869j6` agree
byte-for-byte; `server-2acb_0tu` additionally retains diagnostic enqueue stacks
and matches the same fixture. These runs are under
`artifacts/final-ocean-ruins-reference/` with commands and provenance.

## Reproduction and limits

```powershell
python tools/worldgen/build_final_ocean_fixture.py --preset ocean-ruins --check
python tools/worldgen/build_final_ocean_fixture.py --preset ocean-ruins --trace-decoration --check
python tools/worldgen/build_template_tick_probe.py --check
go test ./internal/world -run '^TestOceanRuinGenerationAgainstFrozenVanillaServer$' -count=1 -v
go test ./internal/world -run '^TestFinalOceanRuinChunksAgainstVanillaServer$' -count=1 -v
```

Published source save: `artifacts/final-ocean-ruins-reference/server-4ufh8myd`.
Fresh independent check: `server-0wuhav2m`; instrumented observer checks:
`server-sb0fpjb0` and `server-ixbkgoi_`. Each run retains commands, console logs,
SHA-256 digests and source-region provenance. Instrumentation observes callbacks
and queues without replacing their algorithms. Validation logs are under
`artifacts/brushable-reference/`.

The final frozen-server gate passes both chunks, including all 196,608 block
states, 24 complete block-entity records, 908 block requests and 814 fluid
requests (`frozen-ticket-order.txt`). The expanded native template capture was
independently regenerated and checked. Its Go comparison passes all 27,792
cases through seven adapters, alongside the direct aquatic and tree callbacks.
The 187 basic tests, three cache/recording race checks and
`go vet ./internal/worldgen ./internal/world` also pass.

Existing runtime acceptance remains green: all 28 controlled boxes at 17
checkpoints, and both no-structure ocean chunks after 100 ticks
(`prior-runtime-regressions.txt`). The complete ocean-ruin, shipwreck and
ruined-portal placement oracles also pass (`aquatic-block-regressions.txt`).
These checks validate the shared-callback changes in their established scopes.

The tool also captures the original chunks after 100 ticks. BrushableBlock's
scheduled reset/falling path is [independently checked](brushable_runtime_1.21.1.md),
and the subsequent [kelp destruction/drop port](kelp_runtime_1.21.1.md) passes both
saved block arrays (`final-ruins-kelp.txt`). With the destruction write flags
corrected, the expanded production-world gate passes both complete 98,304-state
arrays, all 2,515 ordered dispatches and both initial/final live queue pairs
(`live-order-fixed.txt`). The 187 basic tests and world/tickqueue vet pass;
scheduler checkpoint/failure race checks also pass (`live-queue-owner-race.txt`).
All 28 controlled boxes at 17 checkpoints and both no-structure ocean samples
also pass after this correction (`live-order-regressions.txt`).
Full final-world equality is still
unproven: item lifecycle and general shape-destruction loot/events remain
incomplete. The live queue observer above avoids stale disk snapshots.
The existing no-structure
ocean and controlled runtime-arena gates cover different scenarios. Neither these
two frozen samples nor the primitive probe establishes arbitrary-seed equality,
complete entity simulation, or default server clock/ticket behavior.

Background behavior: [Suspicious Sand](https://minecraft.wiki/w/Suspicious_Sand)
and [Suspicious Gravel](https://minecraft.wiki/w/Suspicious_Gravel). Version-specific
callback ordering and constants are taken from the pinned JAR, not inferred from
the current wiki.
