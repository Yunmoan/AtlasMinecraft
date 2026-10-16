# Nether and End feature validation (Java 1.21.1)

The runtime implementations are pure Go. The official Minecraft JAR is used
only to generate development fixtures. Each feature fixture compares the final
block hash, number of writes, return value, and the next random value; matching
the terrain shape alone would miss random-stream and placement-order errors.

| Stage | Official cases | Positive cases | Go checks |
| --- | ---: | ---: | --- |
| Nether forest vegetation and huge fungi | 1,840 | 1,264 | `TestNetherForestFeaturesAgainstVanillaJAR` |
| Twisting/weeping vines, glowstone, basalt pillar/columns, delta and replace blobs | 2,240 | 1,214 | `nether_decoration_test.go` |
| Scattered ore (including ancient debris) | 2,240 | 2,240 returned true | `scattered_ore_test.go` |
| Nether plants, fire and soul fire survival | 7,098 | 2,615 | `nether_survival_test.go` |
| Configured Nether fire/roots/soul-fire patches | 720 | 152 | `TestNetherPatchesAgainstVanillaJAR` |
| Top-level placed Nether patches | 5,760 | 171 | `TestNetherPatchPlacementsAgainstVanillaJAR` |
| End island, platform, gateway, chorus and spikes | 1,296 | 1,216 | `end_features_test.go` |

The placed patch oracle runs the JAR's actual lazy placement streams and shares
the same random stream with the configured feature. Both sides skip only the
biome modifier in these isolated fixtures; this is not a whole-world biome or
neighbor decoration-order proof. Nether generation sea level is 32, its noise
depth is 128, and its dimension build height is 256.

Fire survival uses an extracted 26,684-state `FireBlock.canBurn` table rather
than names guessed to be flammable. This preserves waterlogged-state behavior
and the nonflammable Nether wood family. The scattered ore tests include custom
sizes and target lists, air exposure rejection, zero writes, and Java float
rounding boundaries.

The End fixtures also verify 80 seed-derived spike layout entries, 432 actual
gateway block entity NBT values, and 488 actual End crystal instances (position,
yaw float bits, invulnerability, and beam target). Gateway NBT uses the 1.21.1
`exit_portal` integer array; `ExactTeleport` is omitted when false. End crystal
requests carry entity NBT into generation output and are realized into the
official Anvil `entities` region files. Stable UUIDs and shared runtime entity
IDs survive logout, rejoin and server restart. TCP validation verifies that
spawn and metadata packets follow the visible chunk; independent Anvil readers
verify the saved entities and that pending requests were cleared. Evidence is
`artifacts/generator-validation/structure-entity-binary-v7.txt` and
`structure-entity-anvil-v7.txt`. These checks do not verify AI, entity ticks or
graphical client rendering.

Race checks for these implementations passed. The latest placed patch race
run is `artifacts/worldgen-probes/nether-patch-placement-race-v7.txt`; Nether
decoration, forest, End, survival and scattered ore logs are adjacent files.

These feature fixtures do not establish complete vanilla world equivalence.
Whole-chunk neighbor decoration scheduling, remaining structures, entity
behavior and graphical client rendering require their own verification.
