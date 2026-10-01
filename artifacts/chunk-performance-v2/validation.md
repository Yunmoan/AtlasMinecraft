# Chunk loading and generation performance

Date: 2026-10-01. Windows amd64, AMD Ryzen 9 8945HX, Go 1.27.1,
32 logical CPUs. Runtime remains pure Go and standard library only.

## Changes

- Anvil reads fetch the location entry, record header and target compressed
  payload with `ReadAt`, instead of allocating and reading the entire region.
  Existing size/length bounds and gzip, zlib, raw, LZ4 and external-record
  handling are retained.
- A zero-value region store uses 64 bounded lock stripes. Readers share a
  region lock; replacements serialize within that region. Codec work runs
  outside the file locks, and different regions can progress independently.
  Windows handles close before decompression or a region replacement.
- Zlib writers are reused with `sync.Pool`; resetting to `io.Discard` releases
  references to previous buffers. Region writes still use synced temporary
  files and atomic replacement, preserving neighboring records/timestamps.
- Each region actor has an LRU packet cache bounded by both 64 entries and
  2 MiB. Block revisions and immutable light-snapshot identities invalidate
  stale packets. Public `LoadChunk` returns owned bytes; network sinks share
  immutable buffers. Eviction or later edits never modify queued packets.
- Lighting reuses its state, sky, block and bucket-queue work arrays under the
  engine mutex. Published light arrays remain separately allocated and immutable.
  Empty nibble layers no longer allocate discarded output buffers.
- Density operands are resolved into immutable arrays at compile time,
  removing string-map reads from recursive sampling while preserving the
  arithmetic and interpolation order. Manually constructed probe expressions
  retain their map-based fallback.
- Aquifer centers use a compact 3x3 chunk-local grid, with a general coordinate
  cache for samples outside it. Position seeds and nearest-center tie order
  are unchanged.
- Surface passes borrow immutable rule states internally. Public surface
  samples and final generated chunks detach property maps. Decoration uses
  sparse private block/postprocessing overlays, rather than copying every
  neighboring surface map and postprocessing array for each replay.
- Feature providers, placement modifiers and predicates compile once. A replay
  binds only its private region and biome callback; random streams retain the
  original lazy execution order. Each source computes its decoration seed once.

Generation coverage remains version **4**. Existing saved chunks are loaded
without regeneration. Structures, decorated/special trees, remaining features,
Nether/End and vanilla spawn search remain incomplete. These optimizations do
not establish equivalence of complete decorated worlds to vanilla.

## Reproducible benchmarks

Three repetitions; the table reports medians. A generator operation creates a
fresh seed-1234 generator and produces the entire adjacent 3x3 area including
its required terrain neighborhood. It excludes disk persistence and network
encoding. The populated-region benchmark reads the supplied Java sample without
modifying it. The packet benchmark loads a flat chunk once, then requests it
repeatedly through the real actor, including the public return-buffer copy.
Saved-spawn operations include opening, preparing and closing a 49-chunk flat save.

| Operation | Before | After | Before / after | Allocated bytes before / after |
| --- | ---: | ---: | ---: | ---: |
| Generate 9 chunks, 1 worker | 5.752 s | 3.760 s | 1.53x | 868.235 / 342.164 MB |
| Generate 9 chunks, 4 workers | 2.073 s | 1.313 s | 1.58x | 868.335 / 342.188 MB |
| Read one chunk in a populated region | 1.088 ms | 0.366 ms | 2.98x | 2.966 / 0.327 MB |
| Request an already loaded chunk packet | 65.134 us | 16.030 us | 4.06x | 218.280 / 57.520 kB |
| Reopen/prepare/close 49 saved flat chunks | 370.524 ms | 342.501 ms | 1.08x | 256.539 / 66.641 MB |
| Durable rewrite of one synthetic chunk | 2.597 ms | 2.822 ms | 0.92x | 1.196 / 0.224 MB |

Generation cumulative allocation decreases about **61%**, and warm allocation
count falls from about 3.23 million to 0.69 million per nine-chunk operation.
Saved-area allocation decreases about **74%**. These are cumulative allocated
bytes, **not peak resident memory**. The writer pool reduces allocation; this
small synced-write benchmark shows no latency improvement. Filesystem latency
and host load affect timings, especially single process runs.

Raw evidence: `worldgen-before-final.txt`, `worldgen-after.txt`,
`read-before-final.txt`, `loading-before-final.txt`, `loading-after.txt`,
`packets-before-final.txt`, `packets-after-final.txt`.
The longer packet runs use one second per repetition; the other I/O benchmarks
use 30 iterations. Initial/intermediate measurements are retained separately.

CPU/allocation profiles: `worldgen-before.cpu`, `worldgen-before.mem`,
`worldgen-after.cpu`, `worldgen-after.mem`, `profile-after-top.txt`,
`alloc-after-top.txt`. Profiled timing is excluded from the table. The original
profile attributes 270 MB to decoration neighbor copies and 268 MB to repeated
surface-rule output allocation; sparse overlay writes use about 28 MB in the
new profile. Density/noise evaluation remains the main CPU work.

```powershell
# Current source
go test ./internal/worldgen -run '^$' -bench '^BenchmarkOverworldArea$' -benchtime=1x -count=3
go test ./internal/anvil ./internal/world -run '^$' -bench 'Benchmark(ReadPopulatedRegion|WriteChunk|PrepareSavedSpawn)$' -benchtime=30x -count=3
go test ./internal/world -run '^$' -bench '^BenchmarkLoadedChunkPacket$' -benchtime=1s -count=3

# Preserved baseline source: set the sample path for the region benchmark
$env:ATLAS_BENCH_REGION = 'E:\AtlasMinecraft\vanilla\world\region'
Set-Location E:\AtlasMinecraft\artifacts\chunk-performance-v2\_baseline
go test ./internal/anvil -run '^$' -bench '^BenchmarkReadPopulatedRegion$' -benchtime=30x -count=3
```

`_baseline` contains the source snapshot before these changes with the same
benchmarks. Leading underscore keeps it out of root `go test ./...` traversal.
Individual original files are also preserved in `before` with `.before`
extensions. `source.diff` records the performance implementation changes.

## Runtime and correctness checks

- Full Go tests and `go vet ./...` passed for the optimized build snapshot.
  Checked-in JAR vectors cover noise, biomes, surface, carvers, trees, vegetation,
  placement/random streams, heightmaps and lighting.
- Fresh JDK 21 verification reproduced all **36** reference probes, plus the
  separate column-vegetation set, without rewriting reference files:
  `jar-reference-final.txt`, `jar-columns-final.txt`.
- All three version-4 output hashes pass sequential and concurrent generation.
  Materials, fluid flags, postprocessing, surface states/properties, quart biomes
  and skipped-feature coverage remain identical.
- New tests check concurrent reads/replacements without losing same-region
  writes, invalid record bounds, packet refresh after block/heightmap/light
  changes, cache bounds, caller-owned packet buffers, and sparse cross-border
  overlays against full private chunks. Returned state mutation cannot poison
  shared terrain or feature templates.
- Full race runs passed for `internal/anvil`, `internal/light`, `internal/world`
  and `internal/worldgen`. See `race-final.txt` and the extracted successful
  package results in `race-chunks-final.txt`. GCC 16.2.0 from temporary
  w64devkit 2.10.0 enabled the Windows race detector; the release archive's
  SHA-256 matched the GitHub release digest. No compiler was added to the
  machine's permanent PATH or the runtime build.
- Real server creation/restart smoke passed **53/53** checks on first startup
  and **64/64** after restart, checking all 49 spawn chunks before listening,
  chunk batches, block edits, movement, inventory, position, time and weather.
  `binary-after.txt` and `binary-after/normal-smoke-3995911763` retain the logs.
  In the observed single runs, new normal preparation changed from 9.241 s
  to 5.087 s; restart preparation changed from 371 ms to 266 ms.
- A cold region at chunk **(625045,624993)** with eight workers delivered its
  immediate 3x3 area in **1.995 -> 1.048 s** and all 49 chunks in
  **6.284 -> 2.764 s**. Reload delivery was effectively unchanged:
  **286.792 -> 288.416 ms**. See `remote-before.txt`, `remote-after.txt`.
  The primary gains are initial generation, individual reads and repeated
  packet encoding; full saved-area readiness remains limited by lighting work.
- The actual official Java 1.21.1 server opened and saved both Atlas-created
  normal and flat saves, preserving the edit, quart biomes and **30 generated
  tree/vegetation blocks**. See `vanilla-interoperability.txt`.
- No graphical Minecraft-client validation was performed. Protocol checks and
  JAR references do not validate every seed/location or full world appearance.

## Parallel work and build artifacts

During final checking, independent chat/command edits changed `internal/server`
after the verified build. The user confirmed these were intentional parallel
work. They were preserved. An intermediate full race run failed two outdated
chat/command assertions while all chunk-related packages passed; the parallel
work subsequently updated those tests. Current-source tests and vet are
recorded separately in `tests-current.txt`, `vet-current.txt`, with the server's
race recheck in `race-server-current.txt`.
All these current-source rechecks passed.

The combined build `atlas-current.exe` includes those parallel changes and the
chunk optimizations. Its final binary creation/restart smoke passed **53/53**
and **64/64**, including a saved negative fractional coordinate at
`(34.5,85,-16.5)`: `binary-current-final.txt` and
`binary-current-final/normal-smoke-682314746`.
The test client originally truncated the negative fractional coordinate when
computing the expected chunk, incorrectly expecting Z=-1 instead of Z=-2.
That check now uses `math.Floor(position/16)`; the server's correct coordinates
and the parallel chat/command behavior were preserved.

`atlas-current.exe`: built at 2026-10-01 18:54 (Asia/Shanghai), unsigned Windows
Go binary, no runtime JVM. SHA-256:

```text
5D05F2268204F490C212F6EA9D7A2E6FF0A7CF40EC08E6ED5AD2B3A5A4845BDA
```

The retained **validated optimization snapshot** is `atlas-optimized.exe`, built
at 2026-10-01 18:28 (Asia/Shanghai), before the subsequent chat/command edits.
It is an unsigned Windows Go binary, with no runtime JVM. SHA-256:

```text
03BD100D64298EB2C42A8A3FE200911A6E1EA3AA2BD6A621A6ADD85B7E45B67F
```

The performance source changes are in the shared checkout. The final combined
build is provided separately so the root executable can be managed by the
parallel work. The earlier optimization snapshot keeps the executable used
for the before/after measurement identifiable.
