The Nether and End runtime uses the pure Go `DimensionGenerator`, with the
original dimension noise/surface/biome samplers. Nether carvers use the original
NetherWorldCarver geometry, replacement tag and LegacyRandom stream; the End
has no carvers. Actual build height is 0..256 in both dimensions, while noise
generation depth is 128. Both original dimension types have no skylight.

Placement anchors resolve against generation depth, heightmaps and scans use
build height, and biome filters use each dimension's actual FeatureSorter order.
All 37/37 Nether and 5/5 End placed-feature plans compile and execute in the
production generator, using the JAR-derived order fixture. Complete feature
coverage does not imply complete dimension generation or runtime mechanics.
`count_on_every_layer` preserves vanilla's eager Stream and resamples its count
provider for every loop condition. Its candidate generation must finish before
feature execution consumes more random state.

The final chunk uses 16 block/biome sections and 18 light layers. Immutable
terrain caching and private decoration overlays prevent neighboring generation
from overwriting saved edits. Feature block entities are retained. End crystals
are recorded as pending entity requests and realized by the world bridge into
independent entity Anvil regions, then delivered after terrain.
Unsupported stages remain in the chunk's `AtlasWorldgen.SkippedStages` marker.
This does not claim complete vanilla structures, initial mobs, dragon fight,
portal transitions, tick behavior or original inter-chunk decoration order.

Saved `WorldGenSettings.generate_features` now enables the dimension structure
pipeline. Nether fortress layouts and all fourteen piece placers, all fourteen
Nether fossil templates, Nether ruined portals, and Bastion Jigsaw layouts are generated at their
original structure decoration steps before that step's features. Valid starts,
original piece NBT and neighboring long-array references are saved in Anvil.
Nether fossil `beard_thin` pieces enter actual noise density before surface and
carvers; base columns used to select starts remain uninfluenced, as in vanilla.
Immutable start caches and private owner/clip block caches coalesce concurrent
generation. Large fortresses generate their requested clips lazily, preserving
unprocessed remote chest/spawner flags instead of constructing the entire
fortress footprint when its first chunk is requested. Ruined portals use their
original center clip and retain spread outside the template box; fossil calls
retain the original expanded template clip. Replay still uses a canonical
source order and does not claim arbitrary vanilla neighbor-order parity.

The weighted Nether complex set retains fortress weight 2 and Bastion weight 3.
An accepted Bastion saves and places its complete Jigsaw pieces; its original
start-pool/rotation/template-midpoint biome rejection is checked, so a
basalt-delta rejection can retry fortress. Template processors, clip-local
blocks, loot containers and authored piglins/brutes/hoglins run in pure Go.
Actors retain precise transformed positions, authored persistence and
STRUCTURE finalization, with independent world-region and block-placement
random streams. Finalization currently covers fresh NORMAL worlds with local
difficulty special multiplier zero; enchanted equipment in older/HARD worlds
remains incomplete.

End City starts now use the original placement, biome and raw-noise height
checks. All 20 templates run with their rotations, clipping and overwrite
processors at the original structure step. Treasure markers keep chest loot
tables/seeds; sentry markers generate Shulkers; ship Elytra markers generate
item frames with the original tile position, Facing and displayed item. Full
actor NBT survives DIM1 entity Anvil reopen. Original constructor UUID and
unseeded living-entity yaw are normalized in reference comparisons; this is
not bitwise parity with vanilla's process-global randomness. These actors have
client spawn/metadata encoding, while their gameplay behavior remains incomplete.

The cold 5x5 terrain halo uses bounded parallel workers shared across final
chunk requests. Runtime configuration follows `chunk-generation-workers`;
structure height checks sample raw noise directly and never recursively wait
for a terrain slot. Decoration remains sequential in its canonical source
order. Focused tests compare single-worker and parallel output and exercise
concurrent End City generation with a single terrain slot.
The development convenience constructor leaves structures disabled;
`NewDimensionGeneratorWithOptions` accepts `GenerateStructures` explicitly.

World-side portal primitives now use the stock PortalShape validation and
filling algorithm, destination WorldBorder clamping, and dimension coordinate
scales (Overworld to Nether 1/8, Nether to Overworld 8). `IgniteNetherPortal`
activates an empty valid frame; `NetherPortalSearchTarget` computes the search
anchor. `FindOrCreateNetherPortal` rebuilds the portal POI view from authoritative
chunk states, searches with the original destination radius/distance/Y ordering,
and uses the original PortalForcer spiral, placement checks and fallback geometry.
Actor queries skip sections without portal states; creation uses detached chunk
snapshots. It does not persist the vanilla POI storage or reproduce its tied
record iteration order. `NetherPortalArrival` preserves the source rectangle's
relative player position and rotates yaw/velocity when axes change. It uses
the stock collision search volume, expanded block-shape union and closest free
point, including position-dependent shape offsets. No free point retains the
original candidate, as in vanilla. Border metadata is read at the saved size;
active border lerp/ticks have not been implemented. End exit/credits/dragon
mechanics and portal frame break/update semantics remain incomplete.

`PrepareEndArrival` shares the feature's platform placement helper and rebuilds
the stock arrival geometry: obsidian at Y=48 and cleared air at Y=49..51 around
(100,49,0). Players arrive at (100.5,49,0.5), facing west (yaw 90). Replacement
block drops require the still-incomplete runtime entity system.

`OpenDimension(hub, workers, rootDir, dimension)` and
`OpenStoreDimension(rootDir, dimension)` read root `level.dat` and use the stock
`DIM-1/region` and `DIM1/region` paths. They save their own chunks but do not
rewrite shared world metadata. Player files stay in root `playerdata`, with the
canonical dimension saved in `Dimension` and personal spawn in `SpawnDimension`.

Independent reference checks:

```powershell
python docs/worldgen/verify_reference.py --probe DimensionFeatureOrderProbe
python docs/worldgen/verify_reference.py --probe DimensionPlacementProbe
python docs/worldgen/verify_reference.py --probe NetherCarverProbe
python docs/worldgen/verify_reference.py --probe PortalShapeProbe
python docs/worldgen/verify_reference.py --probe PortalGeometryProbe
python docs/worldgen/verify_reference.py --probe PortalForcerProbe
python docs/worldgen/verify_reference.py --probe PortalRulesProbe
python docs/worldgen/verify_reference.py --probe PortalPositionProbe
python docs/worldgen/verify_reference.py --probe BlockOffsetProbe
python docs/worldgen/verify_reference.py --probe PortalProcessorProbe
python docs/worldgen/verify_structure_reference.py --probe FortressStartsProbe
python docs/worldgen/verify_structure_reference.py --probe FortressBlocksProbe
python docs/worldgen/verify_structure_reference.py --probe NetherFossilStartsProbe
python docs/worldgen/verify_structure_reference.py --probe NetherFossilBlocksProbe
python docs/worldgen/verify_structure_reference.py --probe JigsawStartsProbe
python docs/worldgen/verify_structure_reference.py --probe JigsawBlocksProbe
python docs/worldgen/verify_structure_reference.py --probe JigsawProcessorsProbe
python docs/worldgen/verify_structure_reference.py --probe JigsawBlockEntitiesProbe
python docs/worldgen/verify_structure_reference.py --probe JigsawOverworldBlocksProbe
```

The new fixtures cover 10 nonempty biome schedule groups, 99 placement streams
(positions and subsequent RNG, including variable counts), and 20 complete
Nether carved chunks (every block state, carving mask and counts). Runtime tests
cover deterministic exploration/parallel replay, detached results, 16-section
packets, upper build-height block edits, block light, close/reopen persistence,
metadata isolation, player dimension/respawn and preserved inventory/custom NBT.
Portal references cover 400 shape/frame cases, 96 target/border combinations,
the actual End spawn/arrival coordinates, and 200 ordered platform writes.
Runtime portal tests cover both axes across chunk boundaries, broken/filled
frame rejection, End platform rebuilds and close/reopen persistence.
Additional references cover 54 actual PortalForcer creation calls (origin,
success/failure and the complete ordered write hash), 26,684 state predicates,
30 largest rectangles, 90 relative positions, 36 actual NetherPortalBlock exit
transitions, 376 collision placements, and 132 block offsets. Runtime checks
cover portal creation/reuse, POI rebuild after reopening, cancellation, and
Nether arrivals preserving relative position/yaw/velocity. The original
PortalProcessor fixture covers 1,080 contact/gated/departure steps plus 30 real
ServerPlayer cooldown steps; it verifies the JAR state machine independently
of server contact scheduling. Vanilla player Nether delays are 80 (default)
and 1 (creative), compared before incrementing duration; player cooldown is
10 ticks and refreshes while touching a portal. End's transition delay is 0.

Fortress references cover 56 complete original piece layouts (saved IDs,
orientation, bounds, depth, flags and RNG continuation) and 8,064 actual original
piece postprocess calls. The latter compare every final block, all four
orientations, clipping/supports, fence postprocessing, full chest/spawner NBT,
lava updates, saved flags and decoration RNG continuation. Runtime dimension
tests exercise all four supported Nether families, simultaneous owner/chest/
spawner clips, saved references, detached cache replay, terrain density
influence, actual weighted Bastion rejection, and Anvil close/reopen with
player-edited blocks and shared metadata preserved.

The original Bastion block fixture covers all 167 original elements, 4,008
complete processor chains and 8,016 clipped block/NBT/loot/RNG calls. All 40
built-in processor lists also pass 4,320 isolated original cases, and native
block-entity load/save normalizations cover 83 distinct authored payloads.
Ancient City, Trail Ruins and Trial Chambers add 316 unique elements (including
two lists and missing-resource empty-template fallback), 7,512 complete
processor chains and 15,168 clipped block calls. These calls compare every
final state, block-entity NBT and subsequent decoration random draw. Runtime
Bastion acceptance loads eight clips concurrently, saves all thirty authored
actors from a pinned 88-piece start, and closes/reopens actual DIM-1 block and
entity regions. Replaying pending requests after the simulated entity-write /
block-write crash window preserves UUIDs without adding duplicate actors.
