# Kelp scheduled destruction — Java 1.21.1

Kelp and kelp plants now execute GrowingPlantBlock's delayed support check.
Unsupported plants drop one kelp item, restore their source-water block with
flags 3, and notify neighbors. The implementation is part of the explicit
runtime; generation remains version 18 and random growth is outside this port.

## Evidence and ordering

The [wiki](https://minecraft.wiki/w/Kelp) describes support requirements, the
magma-block exception and one kelp item per broken segment. Exact behavior uses
the pinned Java 1.21.1 server (DataVersion 3955, SHA-1
`59353fb40c36d304f2035d51e7d6e6baa98dc05c`):

- GrowingPlantBlock (`djl`) rechecks survival on its scheduled tick, then calls
  `destroyBlock(pos, true)` if support is gone. It does not remove a supported
  plant just because a previously requested tick exists.
- KelpBlock (`dkc`) and KelpPlantBlock (`dkd`) accept the corresponding head/body
  below them or a sturdy UP face, with magma explicitly vetoed. Shape and
  scheduled callbacks share this survival predicate.
- Level (`dcw`) emits level event 2001, drops resources, replaces the block by
  its fluid legacy state with flags 3 and recursion limit 512, then emits
  BLOCK_DESTROY after a successful write.
- Both pinned `data/minecraft/loot_table/blocks/kelp*.json` tables contain one
  kelp entry and a survives-explosion condition. A support-loss break has no
  explosion radius, so this path yields exactly one item.
- Block.popResource (`dfy`) draws three position offsets before checking
  `doTileDrops`. Even disabled drops consume these draws. Enabled drops then
  draw X/Z velocity; the item's independent RNG draws UUID, bob offset and yaw.
  Item Y subtracts half the native 0.25 height. Pickup delay is 10.

`AtlasPendingKelpBreaks` records the source, expected state and optional full item
NBT before publishing the item. Recovery deduplicates its UUID, restores water
only if the expected kelp still owns the source, and commits source state with
journal removal. A later block edit is preserved. This does not make visual
events or recursive neighbor propagation transactional; retries after a partial
failure may repeat a visual event, and complete crash-atomic neighbor waves
remain outside the current implementation.

## Native and world checks

The native probe calls original GrowingPlantBlock.tick and records its
destruction boundary, then independently invokes original Block.popResource.
It does not replace either algorithm, but it also does not pretend that this
defined level executes the full Level.destroyBlock/neighbor pipeline. That
integration is exercised separately at the world boundary.

All 3,240 cases match: 27 kelp states, 12 supports, five seeds and both drop-rule
values. The comparison includes 810 full item NBT snapshots with UUID and both
RNG continuations. A fresh `--check` independently reproduced the fixture.
Existing falling-block and brushable item references still pass after sharing
their item constructor.

World tests check water restoration, disabled/enabled item creation, random
continuation, full item persistence and an obstructed entity-region write.
Retry after reload preserves the intended UUID and creates no duplicate item.
Additional recovery tests cover an already-published item with the source still
present, already removed, or replaced by a later edit. The 187 basic tests,
world/worldgen `go vet`, targeted world race checks and all 27,792 template
cases pass after the change.

```powershell
python tools/worldgen/build_falling_block_probe.py --kelp --check
go test ./internal/worldgen -run '^TestKelp' -count=1 -v
go test ./internal/world -run '^TestKelp' -count=1 -v
go test ./internal/world -run '^TestFinalOceanRuinChunksAgainstVanillaServer$' -count=1 -v
```

Fixture: `testdata/worldgen/kelp_callbacks_1.21.1.jsonl.gz`.
Logs: `artifacts/brushable-reference/kelp-*.txt`.

## Ocean-ruin result and save boundary

The 100-tick test that previously stopped on `kelp_plant` now completes. Both
98,304-state arrays match the existing official-server saves
(`final-ruins-kelp.txt`). That first check compared block states. The subsequent
[live queue and dispatch reference](ocean_ruin_runtime_reference_1.21.1.md)
observes the original scheduler directly. Item motion, transport, pickup, merging and despawn
remain unimplemented; vibrations and default clock/tickets are also incomplete.

An additional read-only extraction of two independent original worlds found a
save-boundary issue:
the cold chunk has no saved ticks after 100 steps, while the warm chunk still
has exactly its frozen 306 block and 399 fluid entries. The two runs agree
byte-for-byte. Original ChunkMap (`aqb`) skips writing chunks whose `isUnsaved`
flag is false, including during save-all flush. A subsequent read-only agent
observed the native in-memory queues before that guard and recorded every
ServerLevel block/fluid dispatch. The warm chunk did execute all 705 requests;
both live queues are empty after 100 ticks. Its disk queues are the earlier
snapshot, not pending work or a ticket failure. Two fresh instrumented runs
reproduce the same live reference byte-for-byte.

Diagnostic extracts are `final-server-4ufh8myd-*.txt` and
`final-server-0wuhav2m-*.txt` under the same log directory. They were read from
the preserved native worlds, not derived from Go simulation. Those old disk
extracts are not final-queue acceptance fixtures; the separate
`native_ocean_ruin_live_ticks_1.21.1.json` contains the actual live observations.
