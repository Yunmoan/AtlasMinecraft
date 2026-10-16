# Sculk patches and small mushrooms

These implementations execute in Go. The official Minecraft 1.21.1 JAR is
used only to produce the development fixtures.

## Sculk patches

`SculkPatchFeature` (`edi`), `SculkSpreader` (`dmu`), its charge cursor
(`dmu$a`), `SculkBehaviour` (`dmp`), `SculkBlock` (`dmq`) and `SculkVeinBlock`
(`dmv`) supply the algorithm. The port preserves separate worldgen cursors,
spread and growth rounds, cursor delays, 18-neighbor movement, charge decay,
multiface spread ordering, and sensor/shrieker density checks. Both built-in
patch configurations and denser multi-round configurations are checked.

`SculkPatchProbe.java` calls the actual configured features. Its 264 cases
compare final block-state edit hashes, the edited-position count, the next
random long, and postprocessing positions. There are 168 positive cases,
including source/flowing water, bedrock, ordinary stone, ancient-city stone,
leaves/glass, farmland, magma, chunk boundaries, and vertical limits.

Seed `1234`, chunk `(-256, 232)`, now generates 440 sculk blocks, 149 veins,
and 5 catalyst/sensor/shrieker blocks in the integration test. This confirms
runtime routing, not whole-chunk parity with a completed vanilla world.
New catalyst, sensor and shrieker blocks include their default saved NBT.

## Mushroom survival and generation lighting

`MushroomBlock` (`dla`) accepts `mushroom_grow_block` ground immediately.
Other ground needs `getRawBrightness(position, 0) < 13` and a solid-rendering
ground block. Glass/leaves/full collision alone do not satisfy that test.

The FEATURES task runs before INITIALIZE_LIGHT (`ChunkStatusTasks`, `dwb`).
`WorldGenRegion` (`arb`) delegates raw brightness to the server light engine
through `BlockAndTintGetter` (`dbz`). `SkyLightSectionStorage` (`eoy`) returns
15 when the queried sky-light column has no light sections. It does not infer
shadows from the current blocks or their heightmap.

`MushroomFeatureProbe.java` uses the actual `LevelLightEngine`, leaving light
sections absent for fresh regions and queuing real `DataLayer` objects for
the 0/12/13/15 threshold cases. It calls the actual mushroom block and
configured/placed features. The fixture contains 12 brightness cases,
90 survival cases, and 1,144 patch cases. All eight top-level mushroom
placements and both configured patches have positive reference vectors.

Atlas's fresh decoration region therefore reads sky brightness 15 and block
brightness 0. `BindFeatureBrightness` accepts an existing light-engine reader
when a later pipeline supplies previously lit neighboring chunks. The
current deterministic neighboring-decoration replay does not yet supply
those neighbors; lighting-dependent differences in that context remain part
of the outstanding neighboring chunk-stage ordering work.

## Reproducing the focused Go checks

```text
go test ./internal/worldgen -run 'TestMushroomFeatureAgainstVanillaJAR|TestSculkPatchAgainstVanillaJAR|TestDecoratedDeepDarkIncludesSculkPatch|TestVegetationAgainstVanillaJAR' -count=1 -v
go test -race ./internal/worldgen -run 'TestMushroomFeatureAgainstVanillaJAR|TestSculkPatchAgainstVanillaJAR|TestDecoratedDeepDarkIncludesSculkPatch' -count=1 -v
```

To regenerate the Java fixtures, compile both probe files with
`vanilla/server-inner.jar`, its extracted `vanilla/libraries` dependencies,
and the existing `atlasprobe.TreeFeatureProbe` on the classpath. Execute
`atlasprobe.SculkPatchProbe` and save stdout to `sculk_patch_1.21.1.txt`.
Execute `atlasprobe.MushroomFeatureProbe mushroom_features_1.21.1.txt all`
and save stdout to `mushroom_feature_1.21.1.txt`.
