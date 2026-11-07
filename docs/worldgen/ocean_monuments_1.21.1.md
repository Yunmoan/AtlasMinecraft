# Ocean monuments — Java 1.21.1

Generation version 13 enables `minecraft:monument` from the
`minecraft:ocean_monuments` structure set. Runtime generation is Go only.
`jar/server.jar` is a development oracle, pinned by the shared tooling to SHA-1
`59353fb40c36d304f2035d51e7d6e6baa98dc05c` (DataVersion 3955).

## Sources and version boundary

[Minecraft Wiki: Ocean monument](https://minecraft.wiki/w/Ocean_monument)
provides the structure description and a starting point for admission rules.
Exact 1.21.1 behavior comes from the supplied server and its official mappings:
`OceanMonumentStructure` (`emc`), `OceanMonumentPieces` (`emb`), `BiomeSource`
(`dea`) and the original elder-guardian constructor and Mob finalizer.
Current wiki descriptions must not override the pinned version's behavior.

The biome predicate samples the complete inclusive quart-coordinate box within
29 blocks of `(chunkMinX+9, seaLevel, chunkMinZ+9)`, including its vertical
extent, and requires ocean or river biomes. The final generation stub at
`chunkMin+8` uses the base ocean-floor occupied height and requires the monument
deep-ocean tag. The implementation retains Z/X/Y iteration and does not stop
sampling after a rejection; later biome nearest-point ties can depend on the
sampler's preceding result. There is no separately invented radius-16 test.

Layout assembly preserves the original legacy random stream, orientation draw,
room graph shuffle, connectivity pruning, core reservation and room fitter
priority. Saved starts contain the outer `minecraft:omb` piece. Placement
reconstructs the hidden room graph from the world seed, owner chunk and saved
orientation, matching the original load path.

All child rooms, shell sections, foundations, water carving and sponge random
draws are implemented. Water carving preserves existing water, ice, packed ice
and blue ice. The building has three elder-guardian requests, finalized with
the independent world-generation region random. Actor defaults, full health,
persistence, follow-range modifier and handedness are preserved through the
entity Anvil pipeline.

## Independent reference checks

The Java probes invoke original server methods; their output is not generated
by the Go implementation. Checked-in vectors cover:

| Check | Cases | Compared data |
| --- | ---: | --- |
| Layout | 112 | Seven seeds, four origins, four directions; full graph, children, variants, saved NBT and RNG continuation |
| Start admission | 504 | Fixed and natural biomes, base heights, accepted/rejected starts, saved NBT and RNG continuation |
| Block placement | 192 | Full and chunk-clipped geometry, four terrain/water scenarios, all directions; block states, fluid tick positions, guardian positions and RNG continuation |
| Actor finalization | 18 | Six seeds with three consecutive guardians; full saved NBT except nondeterministic UUID, and independent region RNG continuation |

The runtime test uses seed 1234, owner chunk `(13,13)`, and exercises owned
starts, references, multiple clipped chunks, all three actor requests, reverse
replay and mutation isolation. The Anvil test separately checks full guardian
NBT, interrupted-commit replay without duplicates, reopening and unloading.

Later features may change monument blocks. In particular, original
`UnderwaterMagmaFeature` accepts supported non-air/non-water blocks, including
prismarine; it runs at step 6 after the monument at step 4. The integration test
allows this specific later replacement rather than requiring all final-world
blocks to retain their structure-stage state.

```powershell
python tools/worldgen/build_monument_probe.py --probe layout --check
python tools/worldgen/build_monument_probe.py --probe starts --check
python tools/worldgen/build_monument_probe.py --probe blocks --check
python tools/worldgen/build_monument_probe.py --probe actors --check
go test ./internal/worldgen ./internal/world -run '^(TestMonument|TestOverworldMonument)' -count=1
```

These checks establish the specified original method calls and runtime replay.
They do not establish arbitrary neighbor-generation order, complete final-world
block equivalence, fluid simulation/tick parity, graphical-client acceptance or
elder-guardian AI. The other version-13 addition is documented in
[woodland-mansion validation](woodland_mansions_1.21.1.md).
