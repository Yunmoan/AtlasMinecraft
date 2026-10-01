# Java 1.21.1 client compatibility validation

Validated on 2026-10-01, Windows amd64, Go 1.27.1. Target: vanilla Java Edition 1.21.1, protocol 767.

## Changes

- Fixed invisible remote players: runtime entity registry ID 82 is `minecraft:potion`; the player ID is 128. Existing entity-spawn packets used the wrong ID. Player profiles now precede spawns, followed by skin/pose metadata and equipment; tracking waits for play readiness and the observer's terrain.
- Added structured named/RGB colors to command results, TPS values, help, welcome text, chat names and server-list MOTD. Legacy section-sign formats are converted to anonymous-NBT components. Help entries suggest commands when clicked.
- Added typed Brigadier arguments, supported `minecraft:` aliases, `/teleport`, `/save-all [flush]`, and player-name Tab completion. TP accepts online destinations, explicit source players, `@s`/`@p`, relative/local coordinates and vanilla block centering. Block arguments use the complete known 1.21.1 block/state table; time arguments support `t`/`s`/`d` units. Save commands also persist online players.
- Fixed NBT strings to use Java modified UTF-8, including supplementary Unicode characters and NUL. Reading remains compatible with old Atlas ordinary UTF-8.
- Added yellow join/leave broadcasts using vanilla translation keys. Successful joins notify every player in play, including the joining player; later disconnects notify the remaining players regardless of distance. Failed joins do not emit a leave announcement.
- Aligned command-coordinate checks with vanilla bounds and made errors state the allowed range. TP X/Z must be `>= -30000000` and `< 30000000`; Y must be `>= -20000000` and `< 20000000`. Teleport Y is independent of the block build height. The reported `/tp 1000000000 ~ 1000000000` exceeds vanilla bounds and is rejected without changing position or canceling terrain requests.
- Retained `/tps`, `/mspt`, `/status` and `/sysinfo`. Their TPS/MSPT describe the world-clock loop; generation and player handlers run separately, with CPU/memory/queue statistics reported alongside clock metrics.

## Validation

- `go test ./cmd/... ./internal/... -count=1`: passed every production package after the final source changes.
- `go vet ./cmd/... ./internal/...`: passed.
- Two-client TCP tracking test passed: both player spawns decode with literal runtime ID 128, profile and terrain arrive before spawns, movement reaches observers, far teleport removes tracking, player-name TP restores visibility, and disconnect/reconnect removes/recreates entities correctly.
- Three-client TCP presence test passed: a player at `10000000.5 1024 10000000.5` receives the same join/leave announcements as nearby players; joining players receive their own join announcement. Separate tests check that connections before play receive no presence packet.
- Player-name TP tests passed for `/tp Target`, console `/tp Builder Target`, and `/minecraft:teleport @s Target`, preserving destination position and rotation. The exact billion-coordinate command returns the X/Z range while preserving position and terrain requests.
- The supplied vanilla JAR independently decoded the player spawn, command tree, RGB/Chinese/emoji text, TPS text and join/leave translation components. Vanilla Brigadier accepted 30 supported valid commands and rejected 7 invalid commands. Vanilla `Level.isInSpawnableBounds` agreed with 8 boundary cases, including rejection of the billion-coordinate destination.
- Built executable/console/TCP smoke passed **61/61** checks in a separate flat save, including command completion and performance commands. The smoke server stopped cleanly. Logs: [server](smoke/performance-smoke-1295316596/server.txt), [client](smoke/performance-smoke-1295316596/client.txt).

The production package paths intentionally exclude existing incomplete Go source backups under `artifacts`; those archival files were not modified. Runtime tests use isolated saves and ports, without sharing the user's world.

## Limits

No two graphical Minecraft clients were used for visual confirmation. TCP ordering and official codec acceptance are protocol evidence; displayed player models, skins, colors and translated UI still need a real-client check.

Command compatibility covers the supported command subset. Full selector filters, multi-target teleport, teleport facing arguments, weather duration and `setblock destroy` remain unimplemented. Unsupported world-generation stages and other dimensions remain outside this change.

## Build and use

- Server: `E:\AtlasMinecraft\artifacts\client-compatibility\atlas.exe`
- Embedded build time: `2026-10-01T11:15:19Z`.
- Server SHA256: `4AFB2A1BBA518F8E5AD4CA9535BD94A62F7FC867F2F3F4C0769DADA271641FA3`.
- TCP test client: `E:\AtlasMinecraft\artifacts\client-compatibility\atlas-test.exe`.
- Test-client SHA256: `D1F6C943D876D287AA43815390EDA93EC60E2BD47E4C82FA89728E97484FCE5A`.

An existing server at `E:\AtlasMinecraft\atlas.exe` was left running. To use this build with the root configuration and configured save, enter `stop` in the old server console, then run from `E:\AtlasMinecraft`:

```powershell
.\artifacts\client-compatibility\atlas.exe
```

The official JAR is a development reference only and is not required at runtime. Reference fixture reproduction:

```powershell
$env:ATLAS_COMPAT_FIXTURES = Join-Path (Get-Location) 'artifacts/client-compatibility/fixtures'
& "$env:TEMP\atlas-go\go\bin\go.exe" test ./internal/server -run '^TestExportVanillaCompatibilityFixtures$' -count=1
python docs/protocol/verify_player_entity.py artifacts/client-compatibility/fixtures
```
