The Jigsaw template benchmark isolates original template decoding, processor
chains, block-state transforms, clipped placement, block NBT and decoration
random continuation from runtime terrain/neighborhood generation. It uses the
same defined flat-water input as the original block-stage oracle. Each operation
has a fresh writable world; `cold_template` also evicts just its named template
before decoding it again. Registry/transform-table initialization is warmed.
These timings are component measurements, not server chunk throughput.

Four output digests were captured before optimization and pinned in
`TestJigsawPerformanceClipOutput`. They cover Bastion lava-basin, Ancient City
city-center, Trial Chambers chamber-1 and Trail Ruins tower-1 clips, including
final block states, block NBT, authored actor requests and the subsequent random
draw. Existing original-JAR fixtures verify the same code against full processor
and placement calls; a digest alone is not an independent vanilla oracle.

The first optimization removes duplicate block-state normalization/property
copies, discarded Jigsaw block NBT clones, uncapped original-block arrays and
fresh `OnceValues` closures on cache hits. Template decoding then parses source
positions/state indices/NBT once and creates one exactly sized ordered block
array per palette, preserving full-cube/other/entity grouping and stable Y/X/Z
ordering. Connector rotations are cached as immutable relative records; each
start call retains its own origin, shuffled slice and stable priority order.
Saved feature-origin queries traverse NBT directly without a JSON round trip.

Measurements on AMD Ryzen 9 8945HX, Windows amd64, Go benchmark `-benchtime=10x`
(one run each; timings can vary under concurrent development load):

| Component | Before allocated bytes/op | After first two stages bytes/op |
| --- | ---: | ---: |
| Bastion cached template clip | 953,313 | 439,369 |
| Ancient City cached template clip | 3,571,162 | 1,405,345 |
| Trial Chambers cached template clip | 5,286,877 | 1,973,078 |
| Trail Ruins cached template clip | 131,897 | 111,729 |
| Bastion cold template clip | 2,766,812 | 2,022,878 |
| Ancient City cold template clip | 11,705,235 | 7,958,260 |
| Trial Chambers cold template clip | 18,277,713 | 12,479,160 |
| Trail Ruins cold template clip | 495,457 | 427,318 |

Cached template lookup falls from 136 bytes / 5 allocations per call to zero.
Cached processor lookup falls from 176 bytes / 6 allocations to 32 bytes / 1.
The direct saved feature-origin query falls from 1,354 bytes / 16 allocations
to zero bytes / zero allocations; cached start connectors retain only their
private translated/shuffled slice (360 bytes / 4 allocations in the measured
city-center case). Final lookup measurements use 10,000 operations.
The synthetic oracle world's `BlockAt` itself allocates default-state maps;
that test overhead remains identical in these before/after measurements.

Reproduce component output checks and measurements:

```powershell
go test ./internal/worldgen -run '^TestJigsawPerformanceClipOutput$' `
  -bench '^BenchmarkJigsaw(TemplateClip|CachedLookups)$' -benchtime=10x -count=1
```

Artifacts:

- `artifacts/worldgen-probes/jigsaw-template-performance-before-v8.txt`
- `artifacts/worldgen-probes/jigsaw-template-performance-step1-v8.txt`
- `artifacts/worldgen-probes/jigsaw-template-performance-step2-v8.txt`
- `artifacts/worldgen-probes/jigsaw-bastion-focused-race-v8.txt`
- `artifacts/worldgen-probes/jigsaw-cached-lookups-final-v8.txt`
- `artifacts/worldgen-probes/jigsaw-template-performance-validation-v8.txt`
- `artifacts/worldgen-probes/jigsaw-template-performance-race-v8.txt`

The focused pre-performance race suite passed Jigsaw processors, Bastion and
Overworld blocks, cross-chunk feature elements and actual Bastion Anvil restart.
The earlier broad race attempt timed out after 15 minutes while the original
village-start fixture was actively sampling base-height density; that timeout
does not establish either success or a data race for the full start suite.

After all optimizations, the shared-template unit suite passed Jigsaw blocks,
all processor lists, cross-chunk features, parallel Bastion starts, all four
pinned digests, Igloo/Nether Fossil/Ocean Ruin/Shipwreck block-stage references,
fossils, Outpost authored finalization and actual Bastion Anvil restart. The
final focused race suite passed the changed Jigsaw processor/placement/cache
paths and Bastion generation/restart (worldgen 212.370 s; world 20.628 s under
concurrent production-suite verification). These elapsed times are validation
costs, not generation performance measurements.
