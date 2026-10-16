# Village template blocks, processors and terrain projection

All five original Java 1.21.1 village families and their common templates use
the existing pure Go Jigsaw placer. The original JAR supplies development
references only; Atlas does not load it or launch Java at runtime.

## Original assets and call coverage

The village pools contain 62 pools, 478 distinct templates, 16 named processor
lists and 13 feature-pool features. All template entries are legacy single
elements. There are 573 distinct template/processor/projection configurations:
95 desert, 130 plains, 117 savanna, 109 snowy, 101 taiga and 21 common entries.
Of these, 173 use `terrain_matching`, and 400 use `rigid` projection.

`JigsawVillageBlocksProbe` loads every such element through the original
registry codec and original template loader. For each element it checks four
rotations, two positive/negative cross-chunk anchors, flat/sloping/water/lava
worlds and two chunk clips. The reference therefore has 18,336 complete
processor calls and 36,672 block-placement calls. Go compares ordered processed
positions/states/NBT, final block-state edits, saved block-entity NBT, loot seeds
and the next decoration-random long.

The block oracle explicitly disables authored entity instantiation. The 54
authored template actors and 13 feature-pool features have separate oracles and
integration tests; this block reference does not establish their finalization.

## Verified behavior and fixes

Terrain-matching streets apply the original `GravityProcessor` after the
configured processors: each block uses `WORLD_SURFACE_WG(x,z) - 1 + templateY`.
Processors inspect the input world before clipping. Legacy elements append
their air/structure-block filtering after the other processors. Full configured
street, zombie, mossification and farm processing is retained. The five
`farm_*` lists are ordinary original `RuleProcessor` rules, including their
position-seeded crop replacements; there is no separate FarmProcessor class.

The savanna lamp-post template contains an extra closing bracket in its
acacia-fence `final_state`. The original parser accepts the state through the
first closing bracket. Go now follows this behavior, preserving the fence
rather than dropping the block. The original template asset remains unchanged.

The taiga decoration 5 Jigsaw replacement creates a campfire without authored
block NBT. Native block placement creates its default block entity. Go now
records its empty inventory and both four-slot cooking-time arrays after the
block is installed, without consuming decoration randomness. An independent
native constructor/save probe verifies both campfire block types. Replacing an
existing same-type campfire retains its saved cooking progress.

The remaining authored village payloads use the already verified original
load/save codecs. Inventory checks cover all 20 village block-entity block
types, including beds, bells, banners, signs, loot containers, workstations and
campfires. Existing Bastion, other Overworld Jigsaw and outpost block references
continue to pass after the two fixes.

## Reproduce

```powershell
python docs/worldgen/build_jigsaw_data.py --check
python docs/worldgen/build_village_blocks_probe.py --check
python docs/worldgen/build_village_campfire_probe.py --check
& (Join-Path $env:TEMP 'atlas-go/go/bin/go.exe') test ./internal/worldgen -run '^Test(JigsawVillageBlocksAgainstVanillaJAR|JigsawCampfireDefaultsAgainstVanillaJAR|JigsawVillageCampfireReplacementRecordsEntity|VillageTemplateAssetsAndProcessors)$' -count=1 -v -timeout=4m
```

The final focused village plus shared Jigsaw regression run passed in 38.194 s;
the village template vector test took 14.27 s. These are validation timings,
not server throughput benchmarks. Logs and native diagnostics are under
`artifacts/village-worldgen-v11`.

These defined template/terrain calls do not prove complete village worlds or
arbitrary neighboring decoration order. Normal-world start selection, terrain
influence, feature write ownership, actors, Anvil references and reopen checks
are separate runtime acceptance requirements. General entity AI, village POIs,
villager interactions and gameplay ticks remain outside this block port.
