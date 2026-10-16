# Initial Overworld animals, Java 1.21.1

Atlas runs the initial `NaturalSpawner.spawnMobsForChunkGeneration` stage in
pure Go. The supported Overworld species are sheep, pig, chicken, cow, horse,
donkey, armadillo, fox, frog, goat, llama, mooshroom, panda, parrot, polar bear,
rabbit, turtle and wolf. These are generation-time actors; this work does not
implement later natural spawning, AI, movement or entity ticks.

Placement uses original biome spawn weights, group sizes, heightmap selection,
block tags, collision shapes, water obstruction and final-position biome
lookup. Fox and panda have `NO_RESTRICTIONS` placement; parrot uses
`MOTION_BLOCKING`, including leaves. The other supported animals use the
original no-leaves heightmap. `ON_GROUND` lowering comes from the actual
26,684-state `isPathfindable(LAND)` table.

Group finalization preserves horse and biome-derived group variants, baby
states, follow-range modifiers, frog brain cooldowns, fox held items and
subtype-specific fields. Constructor-only egg/scute timers use a separate
position-derived LegacyRandom stream; they do not advance either placement or
world-region random streams.

Official JAR acceptance is recorded in the fixture builders and tests:

- 21,760 actual NaturalSpawner packs, with 2,788 actors across all 18 species,
  including rejection scenarios and both subsequent random states.
- 2,016 subtype finalizations and complete saved-state comparisons, with 504
  following world random states. Constructor-only armadillo `scute_time` is
  outside the finalization-stream comparison.
- 26,684 block states for ground/space checks and placement lowering.
- 2,028 actual subtype metadata codecs, 12 additional entity registry IDs and
  12 frog/wolf variant registry entries.

Fresh normal-world TCP checks use seed 1234 and visit pig, wolf, fox and frog
generation chunks. Two server runs verify stable UUIDs after restart.
Independent Anvil readers verify saved variants, brain cooldown, follow-range
attributes and cleared `AtlasPendingStructureEntities` requests. Evidence:
`artifacts/generator-validation/initial-animal-binary-v8.txt` and
`initial-animal-anvil-v8.txt`. TCP and NBT checks do not establish graphical
client rendering.

Animal brightness checks use the original block opacity and directional face
occlusion rules. A completely unobstructed top-to-position path has exact
skylight 15 and can be answered without constructing a 3x3 lighting engine.
Shaded positions retain the complete neighborhood solver, including emitted
block light. Both paths require the original complete generation neighborhood.
The regression scene compares open sky, stone roofs, leaves, slabs, glass,
water, torches and underground queries with independently assembled complete
lighting. Benchmark evidence is stored in
`artifacts/worldgen-probes/animal-brightness-*-v9.txt` and
`animal-shaded-brightness-*-v9.txt`; these are isolated lighting measurements,
not whole-server generation throughput.

Generated actor constructor snapshots are indexed once as immutable encoded
NBT. Each realization parses an independent object tree, preserving saved
unknown fields and preventing mutations from leaking between actors. The
benchmark logs `actor-realization-*-v9.txt` measure this operation independently
of disk I/O and terrain generation.
