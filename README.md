# Atlas

A from-scratch Minecraft **Java Edition 1.21.1 (protocol 767)** server written in **Go** — no JVM, zero external dependencies, built entirely on the standard library.

## Quick start

```bash
# Build with an embedded UTC build time (PowerShell, Go 1.23+)
./build.ps1

# A plain go build also works, but reports that the build time was not injected
go build -o atlas.exe ./cmd/atlas

# Create a normal Overworld with pure-Go terrain, caves and supported decoration
./atlas.exe -world ./new-world -seed 1234

# Load a COPY of an existing 1.21.1 Java save
./atlas.exe -world ./world-copy

# Create a Java-format superflat world without structures/features
./atlas.exe -world ./flat-world -level-type minecraft:flat -seed 1234

# Import offline-mode operators into ops.json and enable a separate debug log
./atlas.exe -world ./flat-world -ops Alice,Bob -debug

# End-to-end smoke test (drives the full protocol flow against the server)
go build -o atlas-test.exe ./cmd/atlas-test
./atlas.exe -world ./flat-world -level-type minecraft:flat &  # then, in another shell:
./atlas-test.exe        # status→login→config→chunks→chat→dig→commands
# For operator command checks, start Atlas with -ops TestBot, then:
./atlas-test.exe -operator
```

Connect with a **1.21.1** client to `127.0.0.1`. To require authenticated
accounts, set `online-mode=true` in `server.properties` before starting Atlas.

## What works

- Server-list ping (status JSON + ping/pong)
- Offline-mode login or optional online-mode login with RSA, AES/CFB8,
  Mojang hasJoined verification, official UUIDs, and profile properties
  (compression enabled at 256 bytes)
- Full 1.20.2+ configuration phase: brand, **complete 1.21.1 registry data**
  (dimension_type, biomes ×64, damage_type ×47, trim/wolf/painting/banner
  registries — prebuilt as network NBT and `go:embed`ded), feature flags
- Play phase: creative mode, join sequence, position sync w/ teleport confirm
- Existing Java terrain and newly generated normal/structure-free flat chunks streamed
  through chunk batches, with view-distance streaming + unloading
- Configurable spawn-area preparation before accepting players (default 7x7
  chunks), parallel generation workers, deterministic terrain reuse and progress logs
- Chunk jobs follow current player distance and cancel obsolete view requests.
  Teleports cancel pending old-view work before parallel preparation of the
  immediate destination neighborhood; underground ore writes retain height caches.
- Multiplayer: tab list, mutual player entity spawn/despawn by chunk distance,
  relative position/rotation updates, teleport confirmation, arm animation,
  outer skin layer and pose metadata; spawn waits for delivered terrain
- Sky and block lighting use state opacity, emission and face occlusion, with
  cross-chunk propagation, source removal, update packets and Anvil persistence
- Block breaking + creative main/offhand placement, broadcast via block-change
  packets and persisted directly to Anvil
- Player inventory, armor, offhand and selected hotbar slot survive logout and
  restart, including exact network item components. Standard Java Inventory NBT
  exports common components (names, lore, durability, enchantments, potions, food
  and nested containers). Unsupported complex Java imports fail explicitly.
  Player-menu clicks are replayed against server state, with split/merge,
  shift transfer, number-key/offhand swaps and drag distribution.
- Creative interaction range checks, prediction acknowledgments and rejected
  placement correction; solid blocks and player bodies cannot be overwritten.
  Log axes, stairs, slab merging, persistent leaves and double plants have
  initial placement rules. Movement checks block collisions, including half
  slabs, steps and crouching ceilings, using JAR-extracted collision boxes.
- Doors place both halves and open from either half; trapdoors, gates, levers
  and buttons have basic use states. Wall torches, lanterns and switches use
  attachment rules; unsupported attachments/doors are removed, including
  across region boundaries. Fences, bars and panes update their connections.
  Visible players receive mainhand, offhand and armor equipment updates.
- Initial block updates remove unsupported grass/flowers and both halves of
  double plants. Sand/red sand/gravel settle vertically in one durable column
  edit; animated falling entities and drops are not implemented yet.
- Chat (system-chat relay), keep-alive heartbeat with mismatch kick
- Java world save: block edits, world age/daytime/weather, player position,
  inventory and creative flight state survive restart; joining players receive their own
  current chunk first
- Existing Java 1.21.1 (DataVersion 3955) `level.dat`, overworld Anvil
  chunks, and `playerdata/*.dat` can be loaded. Modified chunks are written
  back to `.mca`; player position, flight and inventory are written to vanilla NBT.
- New normal and structure-free flat worlds are created as Java 1.21.1 `level.dat` and
  Anvil saves. The older Atlas hills, `level.json`, and `blocks.log` paths are removed.
- Pure Go 1.21.1 Overworld noise, biome, surface, and carving stages are checked
  against the supplied JAR, including 12 full carved chunks. Feature ordering,
  seeds, definitions, selected placement candidates, ore veins, forest rocks,
  blue ice, and exact motion-blocking/ocean-floor heights are also JAR-checked.
  Normal-world creation and missing-chunk generation now run these stages
  and supported decoration (ores, forest rocks, blue ice, clay/gravel disks,
  selected undecorated trees, grass/ferns, double plants, flowers, dead bushes,
  berry bushes, pumpkins, cactus, sugar cane, waterlilies and huge mushrooms).
  Vegetation includes the original weighted and noise-based providers,
  checked against 4,208 configured/placed JAR cases. Sand/grass disks include
  rule providers and postprocessing, checked against 96 JAR cases. Trees are enabled in taiga, grove, snowy plains,
  windswept hills/forest, wooded badlands, rivers and oceans where their complete
  selector graphs are supported. Structures, decorated/special trees, remaining
  features, and other dimensions are incomplete;
  the complete decorated world is not yet equivalent to vanilla.
- 20 ticks/second world clock and weather controls (clear, rain, thunder)
- Client command tree, chat commands, and console commands: `help`, `list`,
  `where`, `tps`, `mspt`, `status` (`sysinfo`), `seed`, `op`, `deop`, `time`,
  `weather`, `tp`, `setblock`, `save`
- Typed vanilla command arguments, server-side player name completion, and
  namespaced vanilla aliases for supported commands. Help entries can be
  clicked to insert a command; messages support named and RGB text colors.
- `server.properties` for address, port, MOTD, player limit, online mode, view distance,
  advertised simulation distance, world path, and seed; persistent
  vanilla-shaped `ops.json` with UUID-checked operators
- Vanilla-style readable `[HH:mm:ss] [Server thread/INFO]: ...` logs, including
  startup ASCII art, build time, player chat, and issued commands; separate
  debug log and optional packet trace

## World, commands, and logs

Atlas now uses Java 1.21.1 `level.dat`, Anvil regions, and `playerdata` only.
The Atlas hills and their `level.json`/`blocks.log` format are no longer loaded
or created; back up any older Atlas world before using the same directory.
Edits are written to Anvil before acknowledgment. Metadata and online players
save every 30 seconds and on clean shutdown; players also save on disconnect.

Atlas creates `server.properties` and `ops.json` in the server directory if
they are absent. Supported properties are `server-ip`, `server-port`, `motd`,
`max-players`, `online-mode`, `view-distance`, `simulation-distance`, `level-name`,
`level-seed`, and `level-type`, plus Atlas settings `spawn-preload-radius` and
`chunk-generation-workers`; other keys are reported as
ignored at startup. Explicit command-line flags override these properties.
Both distance values currently support 2 through 16 chunks. The view distance
limits chunk streaming and player visibility to the smaller of the server and
client settings. The simulation distance is sent in the join packet; Atlas
does not yet implement vanilla entity and block ticking within that radius.
For a new normal or flat world, a blank `level-seed` chooses a
random seed; numeric values and text values (Java string hash) are accepted.
The seed is saved in `level.dat`. The new flat preset uses the stock four
layers, plains biome, and disabled structures/features. Existing Java worlds
keep their saved seed. The default `minecraft:normal` now creates a noise world
with biomes, surface, caves, ores and supported decoration. New normal saves
explicitly disable structures (`generate_features=false`). `AtlasWorldgen`
metadata records incomplete coverage and each generated chunk lists skipped
placed features. Spawn currently uses column (0,0), not vanilla's spawn search.
Ordinary and fancy oak, birch, spruce and pine tree primitives are JAR-checked.
Forest/plains tree distributions that require beehives, giant trees or other
unimplemented decorators remain skipped; they are not replaced with different
tree distributions. Generation version 4 includes the supported vegetation and disks in new chunks;
existing full chunks keep their saved blocks.

Before opening the listening socket, Atlas loads and saves a square of chunks
around spawn, nearest first. `spawn-preload-radius=3` prepares 49 chunks; the
supported radius is 0..16 chunks and 0 prepares only the center. This preparation
also runs on existing saves, loading their existing chunks without regenerating
them. The radius is independent of view/simulation distance and does not enable
spawn tickets or entity ticking. Startup logs show progress and elapsed time.
`chunk-generation-workers=0` automatically uses half of the available Go CPU
threads, at least one and at most eight. Set a positive value (up to 256) to
choose the worker count explicitly; `-gen 4 -spawn-radius 5` overrides the
properties and prepares 121 chunks. The same worker pool generates chunks
requested while exploring. Generation uses immutable shared terrain, private
decoration overlays and actor-level duplicate-request coalescing; shutdown joins
active generation before returning. Interrupting startup cancels preparation;
already generated chunks remain available on the next start.

Chunk reads fetch only the target Anvil record. Region reads/replacements use
bounded per-region locks; decompression, NBT parsing and compression run outside
those locks. Compression writers and lighting work arrays are reused. Each
region actor caches up to 64 encoded chunk packets and 2 MiB, refreshing them
when blocks or lighting change. Generation precompiles density operands and
feature placement plans, uses compact aquifer grids and sparse decoration
overlays, and materializes only the final requested chunk. Generation version
4 and the checked output hashes are preserved. Measurements and validation:
[chunk loading and generation performance](artifacts/chunk-performance-v2/validation.md).

Gameplay currently remains creative-only. Vertical player motion is still
driven by the vanilla client; server movement validation checks block collisions.
The collision table was extracted at the origin with an empty collision context;
position/entity-dependent shapes (such as bamboo offsets, powder snow and
scaffolding) still require dedicated rules. Beds, remaining block interactions,
fluids, redstone/piston motion, random ticks, mobs, drops, container menus,
crafting and survival are incomplete.
Latest verification: [inventory and interactions](artifacts/inventory-interaction-fix/validation.md).

For an existing Java 1.21.1 save, pass a **copy** of the save directory to
`-world`. Atlas reads `level.dat`, existing overworld `region/*.mca` chunks,
and `playerdata/*.dat`. It writes changed chunk NBT, time/weather, and player
position/rotation/flight/inventory back to those vanilla files, retaining unmodified NBT
tags and keeping a `.dat_old` copy before replacing `level.dat` or player data.
Other dimensions, entities, inventory behavior, and world mechanics are not
yet fully implemented.
Missing chunks using the built-in Overworld noise/biome presets are generated
in Go. Other noise presets, custom enabled datapacks, unfinished existing chunks,
and flat worlds with enabled structures/features are not generated. Existing
full chunks are preserved, and decoration uses private neighboring terrain so
new chunks cannot overwrite saved player edits. The supplied `jar/server.jar`
is a reference only; Atlas never calls
it at runtime. Do not point Atlas at the original `vanilla/world` sample or run
two Atlas processes against one save directory.

Public in-game commands are `/help`, `/list`, and `/where`. Names passed to
`-ops Alice,Bob` can use the world commands; the console can always use them:

```text
/tps
/mspt
/status   (alias: /sysinfo)
/time query [daytime|gametime|day]
/seed
/op <player>
/deop <player>
/time set day|noon|night|midnight|<time>
/time add <time>
/weather clear|rain|thunder
/tp <destination player>
/tp <player> <destination player>
/tp [player] <x> <y> <z>
/setblock <x> <y> <z> <block[state=value]|1.21.1_state_id> [replace|keep]
/save
/save-all [flush]
```

The supported command tree uses vanilla argument types for positions,
players, block states and time, and requests player-name Tab completions from
the server. `/help [command]` displays the permitted commands with clickable
entries. Vanilla command aliases include `/teleport`, `/save-all [flush]`,
and `minecraft:` prefixes for supported vanilla commands. `save`/`save-all`
also save online players before world metadata. Required arguments are not
marked executable prematurely; extra arguments return usage errors.

`/tp <player>` teleports the sender to another online player;
`/tp <player> <destination>` also works from the console. Position forms accept
absolute coordinates, `~` relative coordinates and `^` local coordinates.
Relative coordinates use the command sender's position (console: origin),
and `@s`/`@p` select the sender/nearest online player. Integer absolute X/Z
teleport coordinates center on the block, as in vanilla (`34 85 -17` becomes
`34.5 85 -16.5`); use decimals when an exact edge coordinate is needed.
Teleport destinations retain the vanilla spawnable bounds: X/Z must be
`>= -30000000` and `< 30000000`, and Y must be `>= -20000000` and `< 20000000`.
Teleport Y is not constrained by the block build height. Rejected commands
explain the relevant range and leave the player's position and pending terrain
requests intact; `/tp 1000000000 ~ 1000000000` exceeds the vanilla X/Z bounds.
`/setblock` accepts all known 1.21.1 names and properties, such as
`minecraft:oak_log[axis=x]`, plus numeric state IDs. `keep` only writes to air;
`replace` is the default. `/time query` accepts `daytime`, `gametime`, and
`day`; numeric set/add values accept `t`, `s`, and `d` units. This remains a
supported subset of vanilla commands/selector syntax; full selector filters,
multi-target teleport, teleport facing arguments, weather duration and
`setblock destroy` are not implemented.

Game text uses structured anonymous-NBT components instead of relying on
literal section-sign codes. Success/errors use green/red, labels use aqua/gray,
and TPS values use green (18+), yellow (15..18), or red (below 15). Welcome,
chat names and the server-list MOTD support styled text. Legacy `§0..§f`,
format/reset codes, `§#RRGGBB`, and `§x§R§R§G§G§B§B` are converted to component
styles. Console/file logs stay plain. NBT text uses Java modified UTF-8,
including NUL and supplementary Unicode characters such as emoji.

Player entity spawns use runtime registry ID **128** for Java 1.21.1. The
alphabetically sorted entity-name index 82 refers to a potion in the runtime
registry and must not be used for players. Profiles are sent before entity
spawns, followed by skin/pose metadata and equipment; tracking waits for the
observer's terrain and handles leaving, reconnecting and re-entering view.
Successful joins and subsequent disconnects broadcast yellow vanilla
translation components to all players who have entered play, including the
joining player and observers outside entity visibility range. Failed joins
do not produce a leave announcement, and the client's language controls the
join/leave wording.
Independent development checks can export protocol fixtures with
`ATLAS_COMPAT_FIXTURES` and decode them with
`python docs/protocol/verify_player_entity.py <fixtures-directory>` using the
supplied reference JAR. The reference JAR is never required at server runtime.

Performance commands require operator level 2 or the console. `/tps` reports
actual completed world-clock iterations per wall-clock second over 1, 5, and
15 minutes, capped at the target of 20. Before each window fills, it uses only
the observed runtime; before the first tick it reports that samples are being
collected. Missed iterations are not counted as completed ticks when world
time catches up, and time since the last tick is included even during a stall.
`/mspt` shows clock work average/P95/maximum for the last minute, work exceeding
50ms, the maximum gap between completed loops, last-tick age, and periodic
autosave last/maximum duration. Its work time excludes timer waits and saves;
saves still affect loop gaps and TPS.

`/status` adds uptime, online players, goroutines, Go version/platform,
CPU count/parallelism, heap and Go-managed memory, GC pauses, and cached chunks,
region actors, active generation workers, queued chunk jobs and actor mailbox
depths. On Windows it also reads native process CPU and resident working set;
other platforms currently report those two fields as unavailable. CPU is
averaged since the previous status query (first sample: since server creation);
queries within 250ms reuse the preceding CPU average. 100% means one fully
occupied core, so multicore generation can exceed 100%. Go memory and resident
working set are separate measures. World counters are a best-effort snapshot
without waiting for actors or generation. Clock TPS/MSPT do not measure the
parallel chunk-generation and player-handler workload; inspect CPU, memory
and queues alongside them. The fixed-size history keeps at most 15 minutes
of ticks and sorts latency samples only when queried.

`online-mode=false` is the default. With `online-mode=true`, Atlas verifies the
client's session with Mojang before login and saves players under their official
UUIDs. Existing offline-mode player data and OP entries use different UUIDs
and are not automatically migrated. In online mode, use console `/op <player>`
after that player joins; `-ops` is rejected because it only knows offline
UUIDs. Offline players cannot be granted OP in online mode until they connect.
Console commands can be typed
without the leading slash; console `stop` saves and shuts down the server.
Use `-log-file` and `-debug-log-file` to change log
paths. `-debug` writes detailed events to `logs/debug.log`; `-trace-packets`
adds movement and other high-volume packet logs to that file. Active logs
start fresh on each run, rotate at 10 MiB, and keep three backups.

## Architecture (the interesting part)

```
                    ┌────────────────────────── TCP :25565
                    │  net.Listener (accept loop)
                    ▼
        ┌───────────────────────┐
        │ per-connection         │  read loop (this file's goroutine)
        │ goroutine pair         │  writer loop drains p.Out (buffered 512)
        └──────────┬────────────┘  slow client ⇒ disconnect, never blocks
                   │
             ┌─────▼─────┐
             │    Hub     │  player registry; ViewersOf(x,z) coordinate lookup
             └─────┬─────┘
                   │ messages (mailbox chan)
    ┌──────────────▼───────────────────────────────┐
    │ Region actors — the world is sharded into     │
    │ 32×32-chunk regions, ONE goroutine each.      │
    │ Chunk state is touched only by its owner —    │
    │ no locks on world data, ever.                 │
    └──────────────┬───────────────────────────────┘
                   │ genJob (chan)
        ┌──────────▼──────────┐
        │ chunk-gen worker pool│  pure functions: same input ⇒ same chunk
        └──────────────────────┘
```

- **Region actors**: block edits and chunk requests are messages; the owning
  region goroutine applies them and fans out block-change packets via
  `Hub.ViewersOf`. Cross-region interaction needs no shared state.
- **Backpressure**: every outbound queue is bounded; a lagging client is
  kicked instead of stalling a region actor.
- **Generation boundary**: normal chunks run the verified terrain stages and
  supported features; full decoration and structures are still incomplete.
  Cross-chunk decoration replays neighboring sources in a fixed order using
  an immutable, bounded terrain cache and writes only the requested chunk.

## Repo layout

```
cmd/atlas/         server entry point
cmd/atlas-test/    end-to-end protocol smoke test (status→play)
internal/proto/    VarInt, packet reader/writer, framing+zlib, minimal NBT, packet IDs
internal/server/   handshake/status/login/configuration/play state machines
internal/player/   per-player state + outbound queue
internal/hub/      player registry, viewer lookups, broadcast
internal/world/    chunk codec (palette/heightmap/light), normal/flat generation,
                   region actors, worker pool, Java Anvil/NBT saves, item→block map
internal/logging/  structured rotating logs
internal/blockstate/ shared 1.21.1 block-state codec
internal/worldgen/ vanilla terrain, partial decoration, and JAR reference tests
internal/registry/ embedded 1.21.1 Registry Data payload (prebuilt NBT)
docs/protocol/     protocol data sources + build_registries.py
```

## Regenerating the registry payload

```bash
python docs/protocol/build_registries.py   # reads docs/protocol/mcmeta/, rewrites internal/registry/registries.bin
```

## Not yet (roadmap)

This is not a full vanilla world simulation yet. The provided JAR is a
1.21.1 behavior reference, not a runtime dependency for Atlas.

- [x] Optional online-mode authentication + encryption (AES/CFB8 + Mojang API)
- [ ] Full vanilla server.properties behavior, command dispatcher, permission
      levels, and all game rules (current implementation supports the keys and
      commands listed above)
- [x] Player inventory persistence, common player-menu clicks and equipment sync
- [ ] Complete entity interactions, containers/crafting/drops, survival and health/hunger
- [ ] Exact Java 1.21.1 noise worldgen, biome placement, caves, features,
      and structures (pure Go; required before claiming generated chunks match)
- [ ] Complete vanilla save support: Nether/End, entity regions, block entities
      in client packets, light and heightmap updates, and writing external
      Anvil `.mcc` chunks (LZ4 region reads are supported)
- [ ] Random ticks / redstone; entity AI; automatic weather cycle
- [ ] Persistent regions, region-file I/O off the actor thread

## Notes

- Protocol facts were pinned against machine-readable 1.21.1 data
  (`docs/protocol/`), which caught several traps: `unload_chunk` sends Z
  first; `acknowledge_player_digging` is a bare sequence id since 1.20.3;
  chat components are NBT (not JSON) since 1.20.3; `chunk_batch_finished`
  carries a VarInt.
- Requires a 1.21.1 client; other protocol versions are rejected at login.
