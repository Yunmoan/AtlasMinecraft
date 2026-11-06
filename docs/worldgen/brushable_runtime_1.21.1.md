# Brushable block execution — Java 1.21.1

Suspicious sand and gravel now execute their scheduled brushing reset and
falling behavior. This extends the explicit runtime; it does not change the
generation version (18) or enable the default server clock.

## Source and behavior

The reference is the pinned Java 1.21.1 server, DataVersion 3955, SHA-1
`59353fb40c36d304f2035d51e7d6e6baa98dc05c`. The wiki describes gradual reset
after brushing stops and the exceptional item drop after a long fall:
[Suspicious Sand](https://minecraft.wiki/w/Suspicious_Sand) and
[Brush](https://minecraft.wiki/w/Brush). Exact ordering follows the original
`BrushableBlock` (`dgd`), `BrushableBlockEntity` (`dql`) and `FallingBlockEntity`
(`cjg`) methods, invoked by the development probes.

- The loaded block entity owns brush count, reset deadline, cooldown and hit
  direction. These fields are not saved. A nonzero `dusted` state on reload
  does not reconstruct a brush count.
- At/after the reset deadline, count decreases by two. A change between count
  stages (0, 1–2, 3–5, 6+) writes DUSTED with flags 3. The next reset deadline
  is four ticks later. A remaining count schedules another block tick in two
  ticks; zero count clears deadlines and direction.
- The block tick resets progress before testing support. If it falls, the
  entity retains the state passed into the tick, even when resetting just
  changed DUSTED. Spawn NBT sets `CancelDrop=true` and source removal uses
  the original fluid legacy block.
- A cancelled landing breaks the entity without a block or item. The native
  break callback emits level event 2001 at the floored bounding-box center and
  BLOCK_DESTROY at the continuous center. The center uses the float32 0.98
  entity height. The Go primitive preserves this position; the runtime sends
  the level event, while vibration consumers remain unwired.
- Lifetime/build-height timeout can still create an item despite CancelDrop.
  DropItem and doEntityDrops govern that branch. A failed non-cancelled landing
  emits the break callback only when both drop switches permit it.

Runtime progress lives in the region-owned chunk and is discarded on owning
block replacement or unload. Brushing player input and archaeology loot
extraction are not implemented by this work. Fresh generated entities start
with zero progress; active reset states are independently covered by native
fixtures and the loaded-session test.

## Independent reference coverage

| Gate | Coverage |
| --- | --- |
| Reset and scheduled ticks | 9,240 cases: all eight DUSTED states, seven supports, five build-height boundaries, absent entity or counts 0–9, and times before/at/after reset |
| Spawn results | 4,224 complete falling-entity NBT snapshots and ordered write/schedule/spawn events; world RNG continuation |
| Motion | 992 complete snapshots, eight block states over stone/fence/water/air, ticks 0–30, including 16 cancelled landings |
| Drop decisions | 800 cases over both blocks, five seeds, ten scenes and all doEntityDrops/DropItem/CancelDrop combinations |
| Items | 90 complete item NBT records including UUID and both RNG continuations |
| World ownership | Live reset through scheduled ticks; progress loss after reload; DUSTED preservation; replacement cleanup |
| Persistence | Both suspicious-block falling entities reloaded while airborne at tick 5; original UUID, motion NBT, paused tickets and cancelled landing checked |

The native level records boundary writes without recursively replaying neighbor
updates. It calls original reset, scheduled tick, entity motion and drop methods.
Drop probes seed the original Entity constructor's independent RNG through the
existing development agent; they do not replace movement or item algorithms.
All three brushable fixtures were independently regenerated with `--check`.
Existing ordinary falling callbacks, motion, items, airborne reload and spawn
journal recovery still pass.
The 187 basic tests, targeted world race checks and `go vet` for world/worldgen
also pass. The test-profile preflight discovers all three new reference files;
its four selected brushable reference tests pass (`reference-profile.txt`).

```powershell
python tools/worldgen/build_falling_block_probe.py --brushable --check
python tools/worldgen/build_falling_block_probe.py --brushable --motion --check
python tools/worldgen/build_falling_block_probe.py --brushable --drops --check
go test ./internal/worldgen ./internal/world -run '^TestBrushable' -count=1 -v
```

Fixtures: `testdata/worldgen/brushable_{callbacks,motion,drops}_1.21.1.jsonl.gz`.
Native output: `artifacts/falling-block-reference/brushable-*-stdout.txt`.
Check and Go logs: `artifacts/brushable-reference/` (`runtime-go.txt`,
`world-runtime.txt`, and `*-native-check.txt`).

## Remaining full-world comparison

`TestFinalOceanRuinChunksAgainstVanillaServer` now compares both real structure
chunks after 100 explicit ticks, in addition to the passing frozen-save gate.
The first run stopped at tick 2 in both chunks on `kelp_plant`. The subsequent
[kelp destruction/drop port](kelp_runtime_1.21.1.md) passes both saved block
arrays. Subsequent native live-queue observation distinguishes older disk
snapshots from current pending work. Both ocean-ruin samples now match all
2,515 ordered dispatches and their initial/final live queues. See
`final-ruins-first.txt` for the original failure, `final-ruins-kelp.txt` for the
first block-state pass, and [ocean-ruin fitting](ocean_ruin_runtime_reference_1.21.1.md)
for the expanded acceptance and remaining limits.

Broader falling physics, cross-chunk entity migration, vibration dispatch,
item lifecycle and default clock/ticket integration remain incomplete as
documented in [falling blocks](falling_blocks_1.21.1.md).
