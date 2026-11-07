// Package worldgen implements the supported Java 1.21.1 terrain, biome, feature
// and structure generation stages. Generators share immutable base terrain and
// replay decoration through private overlays. Runtime coordination and storage
// belong to world; standalone random algorithms live in randomsource. Coverage
// of feature plans does not establish complete vanilla world equivalence.
package worldgen
