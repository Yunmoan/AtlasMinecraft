# Spawn preparation and parallel world generation validation

Date: 2026-10-01. Runtime remains pure Go, standard library only.

## Resulting behavior

- Before accepting connections, startup prepares and saves a square around spawn. The default radius is 3 chunks: 7 x 7 = 49 chunks. Requests are ordered by distance; the center finishes first to establish the new save's spawn height.
- `spawn-preload-radius` supports 0..16. Zero prepares only the center. `chunk-generation-workers` supports 0..256; zero selects half the available Go CPU threads, bounded to 1..8. CLI `-spawn-radius` and `-gen` take precedence.
- Startup and exploration share the same generation pool. Duplicate chunk requests coalesce in the owning region actor. Existing full chunks and player edits are loaded without regeneration.
- Preparation requests are bounded by the worker count, use no chunk-packet encoding, and report completed/total/percentage. Cancellation stops further preparation; world shutdown joins active generation, queue submission and region actors before returning, then flushes dirty chunks and metadata.
- Immutable climate search indexing replaces exhaustive scanning of 7,593 biome points. Previous-point and original-order tie preferences are preserved. Noise interpolation reuses the eight corner samples within a cell without changing floating-point interpolation order. Density settings are reused; sampler history stays private to each chunk. Decoration copies touched surface maps; returned chunks are fully detached, including block-state property maps.

## Performance evidence

Machine: AMD Ryzen 9 8945HX, Windows amd64; Go reports 32 logical CPUs. Fixed seed 1234; each benchmark operation constructs a fresh generator and generates the entire adjacent 3 x 3 area, including its required terrain neighborhood. Two repetitions per worker count, without profiling during the table's runs. These are generator timings, excluding disk persistence and network encoding, and are not a Paper comparison.

| Workers | Before, two runs | Final, two runs | Mean improvement |
| --- | --- | --- | --- |
| 1 | 23.63 s / 23.92 s | 6.81 s / 8.76 s | 3.05 x |
| 4 | 15.69 s / 15.70 s | 4.60 s / 4.11 s | 3.61 x |

Steady-state cumulative allocation for the nine-chunk operation fell from about 986.7 MB to 859.9 MB (about 13%). This is allocated bytes over the whole operation, not peak resident memory. Initial runs additionally initialize shared registries. Profiled final runs measured 3.60 s / 3.63 s with four workers; timing varies with machine load.

Raw evidence: [baseline benchmark](worldgen-baseline-bench.txt), [final benchmark](worldgen-final-bench.txt), [baseline output hashes](worldgen-baseline-hashes.txt), `worldgen-baseline.cpu`, `worldgen-final.cpu`. The baseline profile attributes about 49% of CPU time to exhaustive biome lookup. Baseline and final benchmark executables are retained in this directory for reproduction.

```powershell
go test ./internal/worldgen -run '^$' -bench '^BenchmarkOverworldArea$' -benchtime=1x -count=2
```

## Validation

- `go test ./... -timeout=10m` passed after the final interpolation change, including checked-in Java terrain/carver/feature vectors, biome-index comparisons against 12,000 exhaustive queries, equal-distance cases, and existing gameplay tests.
- `go vet ./...` passed.
- Three negative/adjacent chunk outputs retain their pre-optimization SHA-256 hashes, covering materials, fluid updates, surface block states/properties, quart biomes and skipped features. The same hashes pass with concurrent generation of different chunks. Mutating returned material or block-state property data does not poison shared terrain.
- Worker tests hold four real generation-worker goroutines in flight simultaneously, verify repeated/concurrent requests generate a chunk once, propagate generation errors, cancel preparation, and ensure `Close` waits for active generation. Negative-coordinate flat preparation verifies the entire requested area is already saved at readiness, no chunk outside that area was created, and an edit survives another preparation and restart.
- Final binary smoke: [`worldgen-smoke/normal-smoke-147815750`](worldgen-smoke/normal-smoke-147815750). The test checks all 49 saved chunks before the first client can log in, on both initial startup and restart. Four workers override a property requesting one worker. Initial preparation took **8.027 seconds**; restart preparation took **19 milliseconds**. Protocol checks passed **40/40** initially and **44/44** after restart, including block edits, position, weather and time persistence.
- A separate real-process flat-world check verifies `-spawn-radius 2 -gen 3` overrides properties requesting radius 0 / one worker, saves exactly the requested 25-chunk area before listening, and stops cleanly.
- The Windows race detector could not be run: the installed Go toolchain has cgo disabled and no C compiler is available. A temporary compiler download could not complete; the partial archive was removed. Parallel output and worker-overlap tests passed, but they do not establish absence of every possible data race.

No original `vanilla/world` files were modified. This work does not add Paper spawn tickets/entity ticking or complete vanilla decoration/structures, and includes no graphical Minecraft-client validation. Generation coverage remains version 3; structures, decorated/special trees, remaining features and other dimensions remain incomplete.

## Rebuilt artifacts

- `E:\AtlasMinecraft\atlas.exe`, SHA-256 `679B5426B93FF9CC81DDA45AA652AF1825B6931691A89AA856012C3E8923F5DD`.
- `E:\AtlasMinecraft\atlas-test.exe`, SHA-256 `47FC3F9A19970B65ECE44599FA31170FE05B0B905D7923A2A327F8E5C0795CCB`.
