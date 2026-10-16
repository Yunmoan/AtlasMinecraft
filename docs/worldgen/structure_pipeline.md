# Structure generation (Minecraft Java 1.21.1)

The structure code ports original 1.21.1 placement, density adjustment,
scattered/template starts and complete block postprocessing for swamp huts,
desert pyramids, jungle temples, igloos, buried treasure, ruined portals,
shipwrecks, Nether fortresses and Nether fossils. Runtime is pure Go. These verified piece
algorithms do not imply that every structure family or the complete decorated
world matches the original server. Normal-world pipeline integration and Anvil
round-trips have their own tests and coverage boundary.

`build_structure_data.py` extracts 20 structure sets, 34 original structure
definitions and their biome tags from `jar/server.jar`. The extraction checks
version 1.21.1 / DataVersion 3955. Runtime uses the embedded JSON; it does not
launch Java or require an installed JDK.

## Verified components

| Component | Original-JAR vectors | Go API |
| --- | ---: | --- |
| Random-spread candidate chunks, linear / triangular spread, all four frequency reducers and outpost / village exclusion | 55,392 | `LoadStructurePlacementState`, `PotentialChunk`, `AllowsFrequency`, `IsStructureChunk` |
| All five terrain adjustments, rigid pieces, jigsaw junctions, full float kernel and resulting double density | 88,624, exact bits | `NewStructureBeardifier`, `NewStructureBeardifierForChunk`, `Sample` |
| Stronghold concentric rings, per-ring random forks, original preferred biome relocation at Y=0 | 1,536 positions over 6 seeds | `GenerateConcentricRingPositions`, `PopulateOverworldStrongholdRings` |
| Desert pyramid / swamp hut / jungle temple terrain and biome rejection, random orientation, initial bounding boxes and saved piece fields | 1,086 (144 cases, 78 accepted starts) | `GenerateScatteredStructureStart`, `StructureStart.NBT`, `StructurePiece.NBT` |
| Original base-noise occupied height, without full chunk allocation | 384 unique columns from the starts probe | `OverworldNoiseGenerator.BaseOccupiedHeightAt` |
| Swamp hut geometry, vertical alignment, rotations / mirrored stairs, supports, clipping and postprocessing | 240 | `PostProcessSwampHutBlocks` |
| Desert pyramid geometry, four chests, TNT trap, archaeology room, archaeology NBT and both RNG continuation values | 768 | `PostProcessDesertPyramidBlocks` |
| Jungle temple randomized moss geometry, two tripwire / dispenser traps, lever / piston puzzle, two chests, saved flags / loot and RNG continuation | 960 | `PostProcessJungleTempleBlocks` |
| Runtime structure registry decoration step / index and Overworld region random | 34 registry entries and 48 streams with 8 longs each | `OriginalStructureDecorationOrder`, `NewStructureRegionRandom` |
| All original block-state mirror-before-rotation transforms and collision-full-cube sorting classification | 320,208 state transforms | `TransformTemplateBlockState`, `DecodeStructureTemplate` |
| Original integer template transforms, including overflowing pivots | 336 | `TransformTemplatePosition` |
| Precise entity template transforms and original Entity.rotate/mirror yaw arithmetic, exact double/float bits | 480 positions and 144 yaw values | Template entity transformation helpers |
| Igloo template block order, laboratory chance/depth, rotation and saved starts | 713 rows, 48 starts, 203 pieces | `GenerateIglooStructureStart`, `LoadIglooTemplate` |
| Igloo blocks, original shape-update order, full block-entity NBT, loot seeds, rotations, clipped placements and RNG continuation | 2,304 | `PostProcessIglooBlocks` |
| Buried treasure starts, true base ocean floor, biome gate and saved piece | 96 | `GenerateBuriedTreasureStructureStart` |
| Buried treasure descent, neighbor filling, fluid distinctions, chest facing/full NBT and RNG | 576 | `PostProcessBuriedTreasureBlocks` |
| Seven ruined-portal variants, real base columns, placement/biome checks and full starts | 720 starts | `GenerateRuinedPortalStructureStart` |
| Original `ChunkGenerator.createStructures` weighted selection, removal and retry with full start NBT | 1,728 cases | `SelectStructureSetStart` |
| Shipwreck/beached starts, 20 templates, eight palettes, rotations, terrain/fluid protection, single/multiple clips, full NBT and fluid ticks | 288 starts and 8,640 block sequences | `GenerateShipwreckStructureStart`, `PostProcessShipwreckBlocks` |
| Nether fortress layout and complete piece blocks, fences, chest/spawner NBT, lava ticks and RNG | 56 layouts and 8,064 block calls | `GenerateNetherFortressStructureStart`, `PostProcessNetherFortressBlocks` |
| Nether fossil base-column sampling/descent, biome gate, 14 templates/all rotations and expanded clipping | 192 starts and 3,024 block calls | `GenerateNetherFossilStructureStart`, `PostProcessNetherFossilBlocks` |
| Bastion pre-assembly start template, transformed center, biome gate and RNG | 480 cases | `GenerateBastionGenerationStub` |
| Jigsaw pool expansion, junctions, projection, priority, collision, expansion hack and independent trial-chamber aliases | 1,920 original full-start fixtures across ten structure families | `GenerateJigsawStructureStart` |

Ring generation is a world-level prerequisite. The measured six-seed Go test
took 27.43 seconds total, approximately 4.5 seconds per world's 128 relocated
stronghold positions. Cache it once and compute it lazily until a feature
actually needs strongholds. It is not a per-chunk operation.

The expanded single-column height check matched all 384 tested original columns
in 0.51 seconds. It stops at the highest non-air block and does not allocate a
full terrain chunk. These measurements describe focused tests on the current
machine, rather than whole-server throughput or multi-player performance.

## Integration order and limits

1. Initialize structure placements per world and identify potential start
   chunks. Placement does not imply biome or terrain acceptance. Exclusion zones
   deliberately inspect other **placements**, including candidates whose biome
   would later reject the other structure.
2. Generate supported structure starts before base terrain. The current
   scattered-start constructor supports desert pyramids, swamp huts and jungle temples.
   It creates exact initial pieces at Y=64; their final vertical placement
   happens during postprocessing. Weighted structure-set selection retries only
   genuine biome/terrain rejections. An accepted family without a complete
   block generator stops selection, preserving the original weights.
3. Load starts from the complete eight-chunk neighborhood, create references
   with `BuildStructureReferences`, and collect referenced pieces / junctions
   for `NewStructureBeardifierForChunk`. Incomplete neighborhoods must not be
   assumed empty. Piece influence is within 12 blocks horizontally; junction
   bounds are strict as in the original implementation.
4. Sample `cache_all_in_cell(add(router.final_density, beardifier))` before
   aquifers and ore-vein material selection. This is where the original
   `NoiseChunk` adds structure influence. Empty structure inputs contribute
   exactly zero and retain current terrain behavior.
5. Postprocess complete supported pieces in their original structure step
   before that step's configured features. Their per-chunk random seeds,
   clipping, mirrors / rotations, fluid ticks, light postprocessing, chests,
   loot tables, archaeology and entity spawning must all be preserved. The three
   supported piece placers cover blocks and saved loot-table seeds. Hut witch /
   cat and igloo villager/zombie-villager requests join the runtime entity
   pipeline, with TCP and entity-Anvil restart tests. Their AI and interactions
   are separate incomplete stages.
6. Persist valid starts and references in the chunk's `structures` compound.
   Preserve original piece IDs, orientation, NBT tag widths and per-piece saved
   state; never save a start as a completed structure before its blocks can be
   generated and reconstructed correctly.

Igloo templates are extracted unchanged from the official JAR. Template
placement uses the original block grouping and stable Y/X/Z sorting, igloo
processors and shape-update order. Saved sign text/defaults, furnace and brewing
stand contents, chest loot-table NBT and RNG consumption match the official
saved block entities in all 2,304 defined cases. Template entity positions and
yaw arithmetic have independent exact-bit coverage; the laboratory villager
and zombie-villager retain their full NBT through runtime creation and entity
Anvil saving. TCP checks do not establish graphical-client rendering or AI.

The version-10 runtime enables 20 of 29 Overworld-native structure definitions,
all four Nether definitions and the End City definition. Bastions, ancient
cities, trail ruins and trial chambers have integrated starts and block placers.
The independent Jigsaw start API also covers outposts and all five villages;
their start APIs alone do not enable those structures in normal chunks.
`build_jigsaw_data.py` extracts 186 pools, 40 processor lists and 968 referenced
templates. One missing vanilla ancient-city asset explicitly retains the original
empty-template fallback. Strongholds have placement coverage but no piece
assembly. The remaining Overworld families are strongholds, villages,
pillager outposts, ocean monuments and woodland mansions. Mineshafts and End
cities now have runtime starts, pieces, block placement and saved references.

`StructureOrderProbe` uses the original vanilla resource manager and
`RegistryDataLoader.load`, then reads `MappedRegistry.stream`. The unordered
`VanillaRegistries.createLookup` used for data generation is unsuitable for
runtime decoration indices. The original desert-pyramid / jungle-temple /
swamp-hut indices at step 4 are 1 / 4 / 20. The index counts all structures in
the step, including unsupported ones.

The original `WorldGenRegion` random is not an arbitrary salt: derive the
world-seed Xoroshiro positional factory, `fromHashOf` the exact name
`minecraft:worldgen_region_random`, fork positional again and sample at the
center chunk's minimum block X/Z with Y=0. Retain the resulting stream across
all structures and features in that source chunk. Desert-pyramid archaeology
consumes this stream independently of the piece decoration random.

Scattered average-ground alignment samples only the first postprocessed clip
inside the piece box. Vanilla therefore depends on neighbor generation order;
a deterministic replay order needs an explicit end-to-end parity boundary.
The official block probes verify each defined original clip/terrain call,
including random draws made for clipped-out geometry, and do not prove arbitrary
multichunk server generation order matches vanilla.

Cached starts stay immutable; `StructureStart.Clone` deeply copies piece NBT
for worker-private postprocessing. `StructureNBTForChunk` saves only owned
starts and uses the exact original lowercase `starts` and capitalized
`References` fields with deduplicated long arrays.

## Reproduce the official checks

Requires Python 3 and JDK 21 for the probes; Java is development-only.

```powershell
python docs/worldgen/build_structure_data.py --check
python docs/worldgen/verify_structure_reference.py --probe StructurePlacementProbe
python docs/worldgen/verify_structure_reference.py --probe StructureBeardifierProbe
python docs/worldgen/verify_structure_reference.py --probe StructureRingsProbe
python docs/worldgen/verify_structure_reference.py --probe StructureStartsProbe
python docs/worldgen/verify_structure_reference.py --probe StructureHutBlocksProbe
python docs/worldgen/verify_structure_reference.py --probe StructurePyramidBlocksProbe
python docs/worldgen/verify_structure_reference.py --probe StructureJungleBlocksProbe
python docs/worldgen/verify_structure_reference.py --probe StructureOrderProbe
python docs/worldgen/verify_structure_reference.py --probe StructureTransformProbe
python docs/worldgen/verify_structure_reference.py --probe StructureEntityTransformProbe
python docs/worldgen/verify_structure_reference.py --probe IglooStartsProbe
python docs/worldgen/verify_structure_reference.py --probe IglooBlocksProbe
python docs/worldgen/build_structure_order.py --check
python docs/worldgen/build_igloo_data.py --check
python docs/worldgen/build_jigsaw_data.py --check
python docs/worldgen/build_bastion_start_data.py --check
python docs/worldgen/verify_structure_reference.py --probe JigsawStartsProbe
python docs/worldgen/verify_structure_reference.py --probe ShipwreckBlocksProbe
python docs/worldgen/verify_structure_reference.py --probe NetherFossilBlocksProbe
python docs/worldgen/verify_structure_reference.py --probe FortressBlocksProbe
go test ./internal/worldgen -run '^(TestStructure|TestScatteredStructure|TestSwampHut|TestDesertPyramid|TestJungleTemple|TestIgloo|TestTemplate)' -count=1 -v
```

Use `--update` only when deliberately changing oracle input scenarios. The
probes invoke the original placement / `Beardifier` / ring / structure methods;
they do not substitute a second implementation of the algorithms in Java.
