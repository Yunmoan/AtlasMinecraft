# Performance commands validation

Validated on 2026-10-01, Windows amd64, Go 1.27.1.

## Commands

| Command | Output |
| --- | --- |
| `/tps` | Completed world-clock iterations per second over 1m / 5m / 15m |
| `/mspt` | Clock work average / P95 / maximum; slow ticks, loop gaps, last tick age, autosave duration |
| `/status`, `/sysinfo` | Clock metrics, uptime, players, goroutines, native process CPU/working set, Go memory and GC, runtime, chunk generation and actor queues |

Commands require OP level 2+ or the console. Console commands also accept names without `/` and print each result line separately. Client command suggestions and `/help` include the commands for permitted players.

These TPS/MSPT metrics describe the current world-clock loop. Generation and player handlers run independently; CPU/memory/queue counters help assess their workload. A low clock MSPT does not establish that chunk generation is fast. CPU 100% means one fully occupied core; parallel work can exceed 100%. Native CPU/resident memory is currently supported on Windows.

## Validation

- `go test ./cmd/... ./internal/... -count=1`: passed all production packages.
- `go vet ./cmd/... ./internal/...`: passed.
- Quantitative tests cover startup windows, a clock stall with catch-up, 1m/5m/15m windows, history wrap, P95/mean/max/slow work, concurrent snapshots, permissions and system-chat delivery.
- World-stat tests inspect an actively blocked generation worker with another queued job, then confirm cache reuse, completion and shutdown counts.
- `TestPerformanceBinaryConsoleAndProtocol`: passed with the built executables in a separate flat save. All four console commands returned samples and native Windows process counters. Idle TPS was about 19.95 after one second.
- The TCP protocol client passed **53/53** checks, including all four in-game performance commands, login/configuration/chunk streaming, creative interactions, persistence commands and inventory.
- Saved smoke logs: [server](smoke/performance-smoke-1907183020/server.txt), [client](smoke/performance-smoke-1907183020/client.txt).

The initial `go test ./...` and `go vet ./...` attempts failed on existing incomplete `.go` source backups under `artifacts/chunk-performance-v2/before/internal`. Every production package passed; the successful commands above exclude those archival snippets. No archival source was modified.

The protocol check uses the bundled TCP client; a graphical Minecraft client was not used for visual confirmation.

## Build and use

- Executable: `E:\AtlasMinecraft\artifacts\performance\atlas.exe`
- Built with `build.ps1`, embedded build time: `2026-10-01T10:19:08Z`.
- SHA256: `48CB3718219ED215A557F4529994328DE194F34DFFC13144F21D4ADC039731E4`

An existing `E:\AtlasMinecraft\atlas.exe` process was left running. Stop that server cleanly, then run the following from `E:\AtlasMinecraft` to use the same root `server.properties` and configured save:

```powershell
.\artifacts\performance\atlas.exe
```
