# Falling blocks — Java 1.21.1

The explicit runtime now turns updated sand, red sand and gravel into saved
falling entities, advances their motion and places them on supported landing
cells. Player block changes no longer immediately settle an entire column.
An invalid landing or native lifetime/build-height timeout can now replace the
falling entity with its saved item, subject to DropItem and doEntityDrops.
Natural floating terrain remains unchanged until its placement/neighbor callback
requests a tick. This matches the update-triggered behavior described by the
[Minecraft Wiki](https://minecraft.wiki/w/Entities). The wiki's
[falling-block overview](https://minecraft.wiki/w/Falling_Block) also distinguishes
special landing behavior for powder, anvils and archaeological blocks; those
families must not be treated as ordinary sand.

## Pinned source and execution

The oracle is Java 1.21.1, DataVersion 3955, official server SHA-1
`59353fb40c36d304f2035d51e7d6e6baa98dc05c`. Runtime code is pure Go. Java classes
are used only by development probes. Exact behavior follows the pinned methods,
including details where a current wiki covers multiple editions or versions:

- `FallingBlock` (`diu`): placement and neighbor shape schedule two ticks later;
  its scheduled callback tests the block below and minimum build height before
  spawning. `DragonEggBlock` (`dih`) overrides the delay to five ticks.
- `ConcretePowderBlock` (`dhk`): neighboring water can immediately return concrete;
  its DOWN exception and sturdy-face check precede scheduling.
- `FallingBlockEntity` (`cjg`): source removal happens in the block-tick phase,
  using its original fluid's legacy state; constructor state clears waterlogging.
  Motion uses gravity 0.04, drag 0.98 and the native float32-derived 0.98 hitbox.
- `Entity` (`bsr`): collision resolves Y before horizontal axes; fall distance
  accumulates float32 deltas. Fence landing height therefore differs from a
  full-block floor. Falling entities tick after block and fluid queues, under
  a separate entity-ticking ticket predicate.

`World.TickFallingEntities` is explicit and follows `World.TickScheduled` for
each advanced game tick. It does not infer entity permission from the wider
block-ticking halo. Source removal and entity creation use a persistent
`AtlasPendingFallingEntities` intent, with a UUID assigned before the first write.
Recovery deduplicates an already-written entity and preserves a later block edit
at the source. Malformed intents are rejected before removing the source block.
Entity NBT is saved in native entity region files, preserving other entities and
unknown root fields. Empty native regions may be zero-byte files: pinned
`RegionFile` (`dwp`) opens with CREATE and defers the header until a write.
Nonempty truncated headers remain errors.

## Independent checks

| Gate | Compared evidence |
| --- | --- |
| Native callbacks | 3,360 placement/shape/scheduled callbacks, 512 full constructor entity NBT records; 32 states over support and build-height boundaries |
| Native neighbor shapes | 1,152 callbacks from the exhaustive six-face state oracle |
| Native motion | 372 complete per-tick entity NBT snapshots, including six landings; sand/red sand/gravel over stone, fence, water and air |
| Native item drops | 1,200 drop decisions over three terrain blocks, five seeds, ten landing/lifetime/height scenes and all doEntityDrops/DropItem/CancelDrop combinations; 135 complete item NBT records including UUID, plus both RNG continuations |
| Real ServerLevel | 28 isolated boxes × 17 checkpoints through tick 100: 1,949,696 states, block-entity NBT, falling-entity NBT, 5,395 fluid records and 245 block records in native queue order |
| Full ocean chunks | All 217 ordered activation block requests, including 77 sand/gravel requests; both 98,304-state arrays still match after 100 ticks |
| Persistence | Airborne reload at tick 5 retains UUID and every motion NBT field against the native trace; inactive entity tickets pause motion |
| Recovery | Prepared intent, removed source, written entity and later source edit survive reload and repeated recovery without duplication |
| Item commit | Disabled/default/enabled doEntityDrops, precise slab landing position, saved item reload and a failed entity-region write followed by retry/reload |

The real-server boxes add free fall, sand/gravel stacking, a fence landing and
red sand entering water to the existing fluid/plant/leaf/bubble cases. Its
extractor reads original Anvil files; no Go simulation creates expectations.
Falling comparisons remove only independently random UUIDs, retaining native
list order and every other saved field. The separate reload test checks actual
UUID continuity. Other entity families, including plant item drops, are outside
this fixture's entity comparison scope.

The motion probe invokes the actual original entity tick and collision code in
an authored static world; the real-server fixture separately checks interaction
with scheduling, neighbor writes and fluids. Current evidence does not establish
complete entity physics or whole-world equivalence.

```powershell
python tools/worldgen/build_falling_block_probe.py --check
python tools/worldgen/build_falling_block_probe.py --motion --check
python tools/worldgen/build_falling_block_probe.py --drops --check
python tools/worldgen/build_fluid_runtime_fixture.py --check
go test ./internal/worldgen -run '^Test(FallingBlockCallbacks|FallingNeighborShapes|FallingEntityMotion)AgainstVanilla$' -count=1 -v
go test ./internal/worldgen ./internal/world -run '^TestFallingItem' -count=1 -v
go test ./internal/world -run '^Test(FallingEntityReloadAgainstVanilla|FallingSpawnJournalRecovery|FallingSpawnInvalidJournalDoesNotEditWorld|FluidRuntimeBoxesAgainstVanillaServer|FinalOceanChunksAgainstVanillaServer)$' -count=1 -v
```

Fixtures are `falling_block_callbacks_1.21.1.jsonl.gz`,
`falling_block_motion_1.21.1.jsonl.gz`, `falling_block_drops_1.21.1.jsonl.gz` and the expanded
`fluid_runtime_1.21.1.json.gz`. Logs are under
`artifacts/falling-block-reference/`; native arena provenance records the server
properties and full command stream under `artifacts/fluid-runtime-reference/`.
Fresh callback and motion captures reproduce their fixtures. The expanded arena
was published from `server-py8rqh2c` and independently reproduced by
`server-q_elsgr9`. The 187-test basic profile passes; airborne reload, all spawn
recovery stages, malformed-intent rejection and delayed player gravity pass under
Go's race detector. `go vet` passes for `internal/anvil`, `internal/worldgen` and
`internal/world`.

## Item creation and commit boundary

The [item-entity wiki](https://minecraft.wiki/w/Item_entities) describes the
ordinary 10-tick pickup delay and initial health of 5. The pinned
`Entity.spawnAtLocation` and `ItemEntity` (`cjh`) constructors establish the exact
1.21.1 representation: one native stack, Age 0, PickupDelay 10, Health 5 and the
falling entity's precise current position, including fractional slab height.
Two world RNG doubles supply horizontal velocity; Y velocity is 0.2. A fresh
independent LegacyRandom supplies UUID, an unsaved bob offset and saved yaw.
This stream is independent of terrain randomness.

The drop probe controls only the independent seed at Entity's RandomSource
creation, using a development-only ASM agent. It invokes original falling ticks,
Entity.spawnAtLocation, ItemEntity constructors and NBT serialization. The
original world RNG and item RNG continuation are compared as well as all saved
fields; UUID and yaw are not normalized out of this fixture. Native reruns with
`--drops --check` reproduce all 1,200 rows. CancelDrop suppresses a landing item
but does not suppress a lifetime timeout item; DropItem and doEntityDrops still
gate that path. Height decisions use the moved entity's floored block position.

The world adapter stages the item until the owning region can commit one complete
entity list. Falling removal and item insertion therefore share one atomic region
replacement. An obstructed region-file path produces an error while retaining the
old airborne entity; retry and reopen produce one item with stable saved UUID/NBT.
Disabled doEntityDrops consumes neither motion nor item-constructor randomness.
Rule, persistence and failed-commit tests also pass with Go's race detector.

This implements item creation and persistence. Item motion, fluid transport,
merging, pickup, age/despawn progression and client tracking are not implemented
by `TickFallingEntities`, whose scope remains falling entities. The 28-box arena
compares falling entities only; it does not establish item-lifecycle equivalence.

## Remaining work

Entity execution supports ordinary sand/red sand/gravel and now the distinct
[suspicious-sand/gravel reset, break and timeout behavior](brushable_runtime_1.21.1.md).
Specialized powder hardening during motion, anvil damage and dragon eggs remain
unfinished. Cross-chunk migration returns an explicit error. The item
lifecycle after creation is still unfinished. Other-entity collision, contextual or
offset collision shapes, slime/bed bounce, honey/web/bubble interactions,
portals and other inside-block effects still need native probes and ports.

The spawn intent is recoverable, but landing block writes and entity persistence
are not yet one recoverable transaction. Default server clock/ticket integration,
dynamic client spawn/movement packets, broader generation-template update
coverage and arbitrary mixed activation queues remain unfinished. Version 17
adds [shared template delayed callbacks](template_updates_1.21.1.md), with
original edge/finalization and production recorder checks. The two ocean
samples and the closed arenas are bounded checks, not a claim of exact generation
for every seed, dimension or structure. The falling-entity work itself retained
version 16; version 17 marks the subsequent change to generation template queues.
