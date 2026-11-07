# Shipwreck generation and live updates — Java 1.21.1

Generation version 19 restores the fluid request made when a shipwreck template
fills a waterloggable block. The earlier implementation produced the right
blocks and request set but inserted some water requests too late.

The [Minecraft Wiki](https://minecraft.wiki/w/Shipwreck) describes ocean and
beached shipwrecks and their loot chests. Exact callback ordering here comes
from the pinned Java 1.21.1 server, DataVersion 3955, SHA-1
`59353fb40c36d304f2035d51e7d6e6baa98dc05c`; current wiki changes are not assumed
to apply to this version.

## Real-server reference

`build_final_ocean_fixture.py --preset shipwrecks --trace-decoration` creates a
fresh seed-1234 world with structures enabled. It freezes the server before
loading the remote samples, disables random ticks and mob spawning, waits for
each forceload, saves, steps exactly 100 ticks and saves again. The extractor
checks the native structure-start ID in each target chunk, as well as version,
seed, structure setting and FULL status.

| Native chunk | Structure | Frozen states | Block entities | Block ticks | Fluid ticks |
| --- | --- | ---: | ---: | ---: | ---: |
| 4,12 | Shipwreck | 98,304 | 2 chests | 15 | 631 |
| 72,-62 | Beached shipwreck | 98,304 | 1 chest | 49 | 412 |

The fixtures retain complete block arrays, complete typed chest NBT including
loot tables/seeds, and ordered queue lists. The observer also serializes live
queues before ChunkMap's dirty-save guard and records every original scheduled
dispatch. The ocean sample dispatches 15 block and 665 fluid requests; the
beached sample dispatches 94 block and 470 fluid requests over 100 ticks.
No queue sorting or tolerance hides ordering differences.

Source run: `artifacts/final-shipwrecks-reference/server-kfteryhp`.
Independent fresh check: `server-gioztyfu`. A further diagnostic run,
`server-wc12tanm`, records pending fluid lists at activation entry and reproduces
the same published fixtures. Each retains original saves, console output,
commands, compiled observer and provenance. The live fixture SHA-256 is
`5898b9f5ecd6c48e738a0814ef3bd868a8aa0d78c1154a07844bad4999afb262`.

## Cause and correction

Both frozen samples initially matched every block, chest record and request
count. The first ocean queue mismatch was already present before activation:
the first 14 generation requests agreed, but Atlas then requested water at
(64,35,206), while the native next request was at (74,35,207). This excluded
the live scheduler's heap behavior as the cause. Executing the structure clip
directly instead of replaying the canonical cache did not change the failure;
that diagnostic change was removed.

Original `StructureTemplate.placeInWorld` (`ent`) captures the fluid before
writing a template block, loads its block-entity data, then calls the container's
`placeLiquid`. `SimpleWaterloggedBlock` (`dna`) accepts registered source water,
writes WATERLOGGED and immediately schedules water. The later pending-liquid
fill loop uses the same callback, before template edge/finalization passes.

Atlas had folded waterlogging into the first block write without scheduling
that request. The later shape pass eventually requested the same positions,
explaining why a sorted request set and final blocks both passed. The port now
preserves the dry write/NBT/fill sequence, reads the actual previous fluid
state, and schedules immediately after each successful initial or deferred
fill. Canonical structure caching and native direction orders are unchanged.

## Reproduction and limits

```powershell
python tools/worldgen/build_final_ocean_fixture.py --preset shipwrecks --trace-decoration --check
go test ./internal/world -run '^TestShipwreckGenerationAgainstFrozenVanillaServer$' -count=1 -v
go test ./internal/world -run '^TestFinalShipwreckChunksAgainstVanillaServer$' -count=1 -v
go test ./internal/worldgen -run '^TestShipwreckBlocksAgainstVanillaJAR$' -count=1 -v
```

The frozen gate now matches all 196,608 states, three complete chest records,
64 block ticks and 1,043 ordered fluid ticks (`shipwreck-waterlogged-fixed.txt`).
All 8,640 existing primitive cases still match placement, liquid/shape effects,
loot, saved piece NBT, RNG continuation and fluid-request sets. Both production
shipwreck structure pipelines also pass. Logs are under
`artifacts/brushable-reference/`.

The 100-tick gate matches both final block arrays and all 1,244 center-chunk
dispatches. The ocean sample passes its complete live boundaries. The beached
sample still **fails** on the order of its four final flowing-lava requests:
all four positions, types, priorities and delay 21 agree, but the boundary
request at (1167,-42,-992) comes after the three internal requests in Atlas
and before them in the published native run. The strict assertion is retained
(`shipwreck-final-queue-detail.txt`).

Expanded native recording locates the divergence in the neighboring chunks'
tick-91 lava dispatch order. Both native and Atlas deadline tables have capacity
32; enlarging the table is not a justified fix. Native has both forceload
regions in one level, whereas the current Go harness replays each region in a
separate world. Chunk keys (71,-62) and (72,-62) collide in the deadline map.
Global queue membership and activation/insertion history must therefore be
captured and replayed before attributing this discrepancy to a queue algorithm.

Fresh native runs also expose limits in fixture repeatability that the first
three runs did not reveal. `server-m37iko3r` has two different frozen/final
boundary states at (64,36,195) and (64,37,195), while center dispatches and live
queues still match. `server-fisfl58i` matches both block arrays and all center
dispatches but has another final beached queue order. Both runs fail `--check`
and remain preserved; published expectations were not replaced or sorted.
The next reference improvement must record/control the surrounding generation
and activation context, rather than assuming a seed and two target chunks
uniquely determine every boundary write and cross-chunk tie.

The 187 basic tests, world/worldgen vet, and both frozen plus 100-tick ocean-ruin
regressions pass after the version-19 change. This does not turn the remaining
shipwreck queue failure into an acceptance pass.

These two real worlds do not establish arbitrary-seed equivalence. Other
template families need equivalent audits of fluid insertion order and their
surrounding decoration. Item lifecycle, general destruction loot/events and
default clock/ticket integration remain incomplete; these fixtures do not
claim complete entity simulation.
