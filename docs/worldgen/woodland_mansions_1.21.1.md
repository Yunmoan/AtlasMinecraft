# Woodland mansions (Java 1.21.1)

The source is the pinned official `jar/server.jar` (DataVersion 3955, SHA-1
`59353fb40c36d304f2035d51e7d6e6baa98dc05c`). The
[Minecraft Wiki woodland mansion article](https://minecraft.wiki/w/Woodland_mansion)
provides background; current-version biome descriptions are not a substitute
for the 1.21.1 biome tag and implementation. The version here uses dark forest.

`build_mansion_data.py` verifies and extracts all 73 original templates. Runtime
embeds these assets and uses Go; it does not invoke Java. Reference probes call
the actual original methods in a fresh JVM. Expected vectors are never generated
by the Go implementation.

| Boundary | Independent original cases |
| --- | ---: |
| Five floor-plan grids, room flags and RNG continuation | 260 |
| Ordered full layouts, template names, origins, bounds, rotation/mirror and saved piece NBT | 84 layouts / 48,456 pieces |
| Four base-noise corner heights, minimum-height admission, biome gate, complete start and RNG | 224 / 110 accepted |
| Every template, all rotations/mirrors, two random seeds, air/water terrain, full/chunk clips | 7,008 |
| Evoker, vindicator and allay finalized NBT and independent region RNG | 54 plus three defaults |
| Authored block entities loaded and saved by original classes | 9 distinct payloads |
| Cobblestone foundation over uneven solid ground, water and void, full/clipped calls | 144 |

Block vectors compare every written block state, full block-entity NBT, scheduled
fluid-update positions, full marker-actor NBT (excluding UUID), and both random
stream continuation values. The probe binds the original block/fluid tags before
placement. It models fresh generation brightness 15; this is explicit in both
world adapters, rather than estimating light from the roof's geometry.

The placement implementation includes structure-marker filtering, waterlogging,
original shape-update order, chest pair correction after mirrors/clipping, wall
connections, mushroom survival, authored spider spawners and container contents,
chest-marker loot, and foundations. Allay group counts and actor finalization
consume the region RNG inline in marker order. Vindicator equipment consumes
the original enchantment check even when its threshold is zero.

Normal-world integration uses a bounded per-chunk structure cache. The mansion
can extend beyond the smaller scattered structures' canonical region, so a
request places only intersecting template pieces in their original order and
then calls the original-style foundation pass. Saved starts contain all pieces;
generated chunks retain structure references and stable actor replay identities.
The runtime regression pins seed 1234, owner chunk `(8,-127)`, and 535 pieces.
Its six selected clips retain 3,525 dark-oak building blocks and eight actor
requests; reverse replay and mutation isolation pass. The monument and existing
jigsaw/outpost integration regressions also pass after the shared cache change.
The separate Anvil test covers both illagers and multiple allays at one position,
full NBT, interrupted replay, reopening, unloading and mutation isolation.

```powershell
python tools/worldgen/build_mansion_data.py --check
foreach ($probe in 'grid','layout','starts','actors','blocks','foundation','block_entities') {
    python tools/worldgen/build_mansion_probe.py --probe $probe --check
}
go test ./internal/worldgen ./internal/world -run '^(TestMansion|TestOverworldMansion)' -count=1
```

These are method-level original comparisons and local pipeline/persistence
checks. They do not establish whole-world equality, arbitrary neighbor-generation
order, shared RNG consumption across overlapping unrelated structures, all
scheduled block ticks, subsequent fluid simulation, or entity AI. Actor vectors
use fresh NORMAL difficulty with zero local difficulty; aged-world and HARD
equipment finalization are outside this comparison. Final decorated chunks need
their own saved-original-world comparisons.
