# River and ocean terrain regression

These four locations use seed `4511451270791457643`, default Java 1.21.1
Overworld settings and disabled structures. Their complete noise, natural
surface and carved block arrays come from the actual built-in classes in
`jar/server.jar`, through the existing `NoiseChunkProbe`, `SurfaceChunkProbe`
and `CarverChunkProbe`.

| Chunk | Saved-world scene used to select the case |
| --- | --- |
| `-24,61` | Coast with kelp in the decorated save |
| `0,10` | Ocean; the chunk also contains plains biomes |
| `0,43` | River; saved section 3 contains only river biomes |
| `-3,44` | Forest and river edge |

`(-8,55)` is forest at Y64 and is not a river regression case. `(0,10)` has
not been established as a deep-ocean case. The separate diagnostic arrays for
these locations remain under `artifacts/river-ocean-debug`.

The regular tests in `internal/worldgen/river_ocean_regression_test.go` reuse
the existing complete-chunk comparison helpers. Noise checks every state and
fluid-update flag. Surface checks every state. Carvers check every state,
every carving-mask bit, start counts and per-carver mask counts.

These comparisons stop before decoration, saving and fluid ticks. They do
not establish final kelp/vegetation placement, neighboring decoration order,
stored biome arrays or graphical-client rendering. The opt-in diagnostic test
also checks the actual integrated `generateBase` output with structures off.

Regenerate the four official probe outputs into a separate artifact directory:

```powershell
python docs/worldgen/diagnose_terrain_stages.py --seed 4511451270791457643 --chunk -24 61 --chunk 0 10 --chunk 0 43 --chunk -3 44 --output artifacts/river-ocean-debug/terrain-stages-regression
```

After inspecting them, `noise.txt`, `surface.txt` and `carved.txt` correspond to
`river_ocean_noise_1.21.1.txt`, `river_ocean_surface_1.21.1.txt` and
`river_ocean_carver_1.21.1.txt`, respectively. The helper never changes the
checked-in fixtures itself. Java remains a development oracle only.

Run just the three ordinary regression tests:

```powershell
& (Join-Path $env:TEMP 'atlas-go/go/bin/go.exe') test ./internal/worldgen -run '^TestRiverOcean(Noise|Surface|Carvers)AgainstVanillaJAR$' -count=1 -v
```
