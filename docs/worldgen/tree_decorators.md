The runtime remains pure Go. The Java probes call the official Minecraft Java
1.21.1 server JAR only to produce and verify reference vectors.

`TreeDecoratorProbe` covers all six decorator types, their block states,
ordering and subsequent worldgen random draws: trunk vines, hanging leaf vines,
cocoa, nest placement, altered ground and blocks attached to leaves. Its 420
cases include obstructed positions, uneven ground, negative coordinates,
world-border coordinates and trees without leaves. Nest occupants are verified
by their entity ID, ticks in hive and minimum occupation time (600 ticks).

`TreeFeatureProbe` also verifies full configured-tree output after leaf-distance
propagation. The additional fixture sets contain 840 decorated, 600 giant,
240 branching, 240 cherry and 304 mangrove cases, alongside the existing
600 base cases.
The scenarios include negative coordinates, cross-chunk positions, preexisting
logs and leaves, submerged trees, clipped trees and build limits.

Reproduce using JDK 21 and the official JAR in `jar/server.jar`:

```powershell
python docs/worldgen/verify_reference.py --probe TreeDecoratorProbe
python docs/worldgen/verify_reference.py --probe TreeFeatureProbe --tree-set decorated
python docs/worldgen/verify_reference.py --probe TreeFeatureProbe --tree-set giant
python docs/worldgen/verify_reference.py --probe TreeFeatureProbe --tree-set branching
python docs/worldgen/verify_reference.py --probe TreeFeatureProbe --tree-set cherry
python docs/worldgen/verify_reference.py --probe TreeFeatureProbe --tree-set mangrove
```

Two official implementation details affect exact reproduction:

* `BeehiveDecorator.place` uses `Collections.shuffle(candidates)`, whose random
  source is process global and independent of the world seed. The probes reset
  this separate Java RNG to `worldgenSeed ^ 0x123456789abcdef0` for each test
  case. The Go decorator accepts the corresponding independent shuffle callback.
  Runtime dispatch freezes this independent choice using a random per-world salt
  persisted in `AtlasWorldgen.BeehiveShuffleSalt`. SHA-256 hashes that salt,
  world seed, source chunk, feature step/index, configured-feature name, tree
  position and complete RNG snapshot to seed a separate LegacyRandom shuffle.
  This source differs from vanilla's process-global RNG. It consumes no worldgen
  random calls and remains stable across neighboring replays, workers and restarts,
  without an unbounded cache.
* The 1.21.1 cherry bootstrap constructor uses a corner-hole chance of 0.5,
  while the JAR's configured-feature JSON stores 0.25. The foliage CODEC's
  corner-hole getter erroneously serializes the wide-bottom chance. Dedicated
  server `WorldLoader` loads its worldgen registries through
  `RegistryDataLoader.loadContentsFromManager`, which parses the JSON resource
  with `ConfiguredFeature.DIRECT_CODEC`. The probes therefore decode that
  actual resource, matching the runtime's extracted configuration, and do not
  replace it with the different bootstrap value.

All 31 vanilla `minecraft:tree` configured definitions can now be decoded and
their trunk, foliage, root and decorator algorithms placed. Mangrove roots are
covered by both types of configured mangrove trees, including source/flowing
water, muddy roots, moss carpets, hanging propagules and simulated-root failure.
Bee-decorated configurations, including mangrove trees, are enabled using that
frozen independent shuffle. Independent replay tests check final tree blocks,
bee NBT, worldgen RNG consumption and generator recreation across eight workers.
