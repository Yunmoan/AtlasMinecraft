# Ordered generation postprocessing (Java 1.21.1)

Generation version 16 retains each mark as an append event. Section lists keep
the original order and repeated positions, even when later generation replaces
the marked block. Existing saved chunks are not regenerated. The runtime remains
pure Go; Java is used only for independent reference capture.

## Pinned original behavior

The official 1.21.1 server JAR (`59353fb40c36d304f2035d51e7d6e6baa98dc05c`)
and its mappings define these details:

- `ProtoChunk.markPosForPostprocessing` (`dvs.e(jd)`) checks build height and
  appends `X | Y<<4 | Z<<8` to that section's `ShortList`. It does not deduplicate.
- `NoiseBasedChunkGenerator.doFill` (`dzb.a(...)`) marks nonempty fluid states
  while traversing cell X, cell Z, descending cell/local Y, local X, local Z.
- `WorldCarver.carveBlock` (`ebe.a(...)`) marks the carved fluid when the aquifer's
  current scheduling flag is true. Below the lava floor it bypasses aquifer
  computation and retains the previous flag. A restored surface containing fluid
  is also marked, independently of that flag.
- `WorldGenRegion.setBlock` marks states whose native `hasPostProcess` predicate
  is true after every accepted write, including unchanged states. Explicit
  feature marks remain separate events.
- `LevelChunk.postProcessGeneration` (`dvi.H()`) traverses sections and their
  lists in order, runs immediate fluid ticks and neighbor-shape updates, then
  clears the lists before unpacking pending generation ticks.

`NoiseChunkResult.PostProcessingOrder` is authoritative for serialization.
Boolean membership views remain only for existing diagnostics. Borrowed feature
overlays and structure records carry ordered logs; snapshot installation and
terrain cloning own their copies. Structure replay restores final block states
without running the automatic marking predicate again, then appends the recorded
marks. This also preserves marks for overwritten intermediate states.

## Independent checks

`build_postprocessing_probe.py` executes original methods and reads native
`ProtoChunk` lists. It never uses Atlas output as expected data:

- 12 chunks at noise and carved stages: **24 lists, 44,678 entries, 686 repeats**.
  The noise probe invokes original `doFill` with the original empty-beardifier
  NoiseChunk. Structure-manager construction alone is bypassed.
- 21 cave feature families × 8 seeds × 64 scenes: **10,752 placements,
  56,449 entries, 8,698 repeats**. Full changed-block hashes, placement result and
  next RNG value are checked alongside every per-chunk/per-section list.
- Cache tests cover private and borrowed terrain, delegated structure recording,
  unchanged writes, intermediate overwrites, clipped replay, out-of-height marks,
  and independent snapshots. An Anvil write/read/edit check retains exact ordered
  lists even when their final blocks are air.

Fixtures: `terrain_postprocessing_1.21.1.jsonl.gz` and
`features_postprocessing_1.21.1.jsonl.gz` under `testdata/worldgen/`.
Capture and test logs are under `artifacts/postprocessing-reference/`.
The combined native noise-fill/carver probe also reproduces all 12 existing
full carved block arrays and masks. Complete ocean arrays still match the
frozen save and the original server after 100 ticks; 133 frozen generation
block requests and all 24 runtime scenes at 17 checkpoints pass. The basic
profile has 186 passing tests, with separate cache/concurrent generation race
checks and `go vet` for world generation and world runtime.

```powershell
python tools/worldgen/build_postprocessing_probe.py --check
go test ./internal/worldgen -run 'Test(TerrainPostprocessing|FeaturePostprocessing|Postprocessing)' -count=1 -v
go test ./internal/world -run '^TestGeneratedPostprocessingAnvilRetainsOrderAndDuplicates$' -count=1
```

These stage and primitive probes do not establish complete mixed activation queue
equivalence for arbitrary fully decorated worlds. The subsequent
[falling-block runtime](falling_blocks_1.21.1.md) now reproduces all 217 ordered
activation requests in the two ocean fixtures, including 77 sand/gravel requests.
Specialized entity behavior, other delayed families and default server clock/ticket
integration remain unfinished; the generator must not be advertised as exact.

The falling-block implementation retains update-triggered behavior:
the [Minecraft Wiki](https://minecraft.wiki/w/Entities) notes that sand/gravel
above generated caves can remain suspended until a block update. The pinned
`FallingBlock` (`diu`) schedules its check after two ticks from placement or
neighbor-shape updates, then creates a `FallingBlockEntity` only when its
support/build-height predicate permits. This is not an unconditional terrain
settling pass.
