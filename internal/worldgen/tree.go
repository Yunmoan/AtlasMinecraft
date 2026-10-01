package worldgen

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

//go:embed tree_tags.json
var treeTagData []byte

var treeTags = func() map[string]map[string]bool {
	var data map[string][]string
	if err := json.Unmarshal(treeTagData, &data); err != nil {
		panic(err)
	}
	tags := make(map[string]map[string]bool)
	for name, blocks := range data {
		tags[name] = make(map[string]bool)
		for _, block := range blocks {
			tags[name][block] = true
		}
	}
	return tags
}()

// TreeConfiguration supports the vanilla straight/fancy trunk and
// blob/fancy/pine/spruce foliage placers with simple state providers. Decorated
// or rooted trees are rejected rather than silently simplified.
type TreeConfiguration struct {
	Trunk, Foliage, Dirt                         SurfaceBlockState
	TrunkType, FoliageType                       string
	BaseHeight, HeightRandA, HeightRandB         int32
	Limit, LowerSize, UpperSize                  int32
	MinClippedHeight                             *int32
	ForceDirt, IgnoreVines                       bool
	radius, offset, foliageHeight, trunkHeight   featureIntProvider
	bendLength                                   featureIntProvider
	minLeaves, leafAttempts                      int32
	trunkProvider, foliageProvider, dirtProvider featureStateProvider
}

func LoadTreeConfiguration(name string) (TreeConfiguration, error) {
	definition, err := LoadConfiguredFeatureDefinition(name)
	if err != nil {
		return TreeConfiguration{}, err
	}
	if definition.Type != "minecraft:tree" {
		return TreeConfiguration{}, fmt.Errorf("%s is not a tree", name)
	}
	var raw struct {
		TrunkPlacer struct {
			Type      string          `json:"type"`
			Base      int32           `json:"base_height"`
			A         int32           `json:"height_rand_a"`
			B         int32           `json:"height_rand_b"`
			MinLeaves int32           `json:"min_height_for_leaves"`
			Bend      json.RawMessage `json:"bend_length"`
		} `json:"trunk_placer"`
		FoliagePlacer struct {
			Type                   string `json:"type"`
			Radius, Offset, Height json.RawMessage
			TrunkHeight            json.RawMessage `json:"trunk_height"`
			FoliageHeight          json.RawMessage `json:"foliage_height"`
			Attempts               int32           `json:"leaf_placement_attempts"`
		} `json:"foliage_placer"`
		Size struct {
			Type    string `json:"type"`
			Limit   int32  `json:"limit"`
			Lower   int32  `json:"lower_size"`
			Upper   int32  `json:"upper_size"`
			Clipped *int32 `json:"min_clipped_height"`
		} `json:"minimum_size"`
		Decorators []json.RawMessage `json:"decorators"`
		Root       json.RawMessage   `json:"root_placer"`
		Force      bool              `json:"force_dirt"`
		Ignore     bool              `json:"ignore_vines"`
	}
	// Explicit names for provider fields (JSON underscores are significant).
	var providers struct {
		Trunk   json.RawMessage `json:"trunk_provider"`
		Foliage json.RawMessage `json:"foliage_provider"`
		Dirt    json.RawMessage `json:"dirt_provider"`
	}
	if err := json.Unmarshal(definition.Config, &raw); err != nil {
		return TreeConfiguration{}, err
	}
	if err := json.Unmarshal(definition.Config, &providers); err != nil {
		return TreeConfiguration{}, err
	}
	config := TreeConfiguration{TrunkType: raw.TrunkPlacer.Type, FoliageType: raw.FoliagePlacer.Type,
		BaseHeight: raw.TrunkPlacer.Base, HeightRandA: raw.TrunkPlacer.A, HeightRandB: raw.TrunkPlacer.B,
		Limit: raw.Size.Limit, LowerSize: raw.Size.Lower, UpperSize: raw.Size.Upper,
		MinClippedHeight: raw.Size.Clipped, ForceDirt: raw.Force, IgnoreVines: raw.Ignore}
	if len(raw.Decorators) != 0 || (len(raw.Root) != 0 && string(raw.Root) != "null") {
		return config, fmt.Errorf("tree %s has unsupported decorators or roots", name)
	}
	if config.TrunkType != "minecraft:straight_trunk_placer" && config.TrunkType != "minecraft:fancy_trunk_placer" && config.TrunkType != "minecraft:bending_trunk_placer" {
		return config, fmt.Errorf("unsupported trunk placer %s", config.TrunkType)
	}
	if raw.Size.Type != "minecraft:two_layers_feature_size" || config.BaseHeight < 0 || config.BaseHeight > 32 ||
		config.HeightRandA < 0 || config.HeightRandA > 24 || config.HeightRandB < 0 || config.HeightRandB > 24 ||
		config.Limit < 0 || config.LowerSize < 0 || config.LowerSize > 16 || config.UpperSize < 0 || config.UpperSize > 16 {
		return config, fmt.Errorf("unsupported tree size for %s", name)
	}
	stateProvider := func(data json.RawMessage) (SurfaceBlockState, error) {
		var provider struct {
			Type  string            `json:"type"`
			State SurfaceBlockState `json:"state"`
		}
		if err := json.Unmarshal(data, &provider); err != nil {
			return SurfaceBlockState{}, err
		}
		if provider.Type == "minecraft:weighted_state_provider" {
			return SurfaceBlockState{}, nil
		}
		if provider.Type != "minecraft:simple_state_provider" || provider.State.Name == "" {
			return SurfaceBlockState{}, fmt.Errorf("unsupported tree state provider %s", data)
		}
		return provider.State, nil
	}
	if config.Trunk, err = stateProvider(providers.Trunk); err != nil {
		return config, err
	}
	if config.Foliage, err = stateProvider(providers.Foliage); err != nil {
		return config, err
	}
	if config.Dirt, err = stateProvider(providers.Dirt); err != nil {
		return config, err
	}
	config.trunkProvider, err = decodeFeatureStateProvider(providers.Trunk, func(SurfaceBlockState) error { return nil })
	if err != nil {
		return config, err
	}
	config.dirtProvider, err = decodeFeatureStateProvider(providers.Dirt, func(SurfaceBlockState) error { return nil })
	if err != nil {
		return config, err
	}
	config.foliageProvider, err = decodeFeatureStateProvider(providers.Foliage, func(SurfaceBlockState) error { return nil })
	if err != nil {
		return config, err
	}
	if config.TrunkType == "minecraft:bending_trunk_placer" {
		config.bendLength, err = decodeFeatureIntProvider(raw.TrunkPlacer.Bend, 1, 64)
		if err != nil {
			return config, err
		}
		config.minLeaves = raw.TrunkPlacer.MinLeaves
	}
	if config.radius, err = decodeFeatureIntProvider(raw.FoliagePlacer.Radius, 0, 16); err != nil {
		return config, err
	}
	if config.offset, err = decodeFeatureIntProvider(raw.FoliagePlacer.Offset, 0, 16); err != nil {
		return config, err
	}
	switch config.FoliageType {
	case "minecraft:random_spread_foliage_placer":
		config.foliageHeight, err = decodeFeatureIntProvider(raw.FoliagePlacer.FoliageHeight, 1, 512)
		config.leafAttempts = raw.FoliagePlacer.Attempts
		if config.leafAttempts < 0 || config.leafAttempts > 256 {
			return config, fmt.Errorf("invalid random foliage attempts")
		}
	case "minecraft:blob_foliage_placer", "minecraft:fancy_foliage_placer", "minecraft:pine_foliage_placer":
		config.foliageHeight, err = decodeFeatureIntProvider(raw.FoliagePlacer.Height, 0, 24)
	case "minecraft:spruce_foliage_placer":
		config.trunkHeight, err = decodeFeatureIntProvider(raw.FoliagePlacer.TrunkHeight, 0, 24)
	default:
		err = fmt.Errorf("unsupported foliage placer %s", config.FoliageType)
	}
	return config, err
}

func treeReplaceable(state SurfaceBlockState) bool {
	return state.Name == "minecraft:air" || state.Name == "minecraft:cave_air" || state.Name == "minecraft:void_air" || treeTags["replaceable_by_trees"][state.Name]
}

func stateWithProperty(state SurfaceBlockState, key, value string) SurfaceBlockState {
	properties := make(map[string]string, len(state.Properties)+1)
	for k, v := range state.Properties {
		properties[k] = v
	}
	properties[key] = value
	state.Properties = properties
	return state
}

type treePlacement struct {
	world          FeatureBlockWorld
	random         *WorldgenRandom
	config         TreeConfiguration
	trunks, leaves map[FeaturePosition]bool
	trunkOrder     javaPositionSet
	err            error
}

func (tree *treePlacement) block(position FeaturePosition) SurfaceBlockState {
	if tree.err != nil {
		return SurfaceBlockState{}
	}
	state, err := tree.world.BlockAt(position)
	if err != nil {
		tree.err = err
	}
	return state
}

func (tree *treePlacement) set(position FeaturePosition, state SurfaceBlockState, set map[FeaturePosition]bool) {
	if tree.err != nil {
		return
	}
	// Vanilla's consumer records attempted positions even if setBlock refuses
	// a foliage row above the build limit. The trunk-height guard does not cover
	// every foliage offset; rejected writes must not abort chunk generation.
	if tree.world.InWorld(position) {
		if err := tree.world.SetBlock(position, state); err != nil {
			tree.err = err
			return
		}
	}
	set[position] = true
}

func treeOffset(position FeaturePosition, x, y, z int32) FeaturePosition {
	return FeaturePosition{position.X + x, position.Y + y, position.Z + z}
}

func abs32(value int32) int32 {
	if value < 0 {
		return -value
	}
	return value
}

func (tree *treePlacement) log(position FeaturePosition, axis string) {
	if treeReplaceable(tree.block(position)) {
		state := tree.config.Trunk
		if tree.config.trunkProvider != nil {
			state = tree.config.trunkProvider(tree.random, position)
		}
		tree.set(position, stateWithProperty(state, "axis", axis), tree.trunks)
		tree.trunkOrder.add(position)
	}
}

// PlaceTreeFeature ports TreeFeature's configuration path and its final
// bounded leaf-distance propagation. Sapling survival is a placement modifier,
// not a configured tree's ground check (the trunk placer may replace stone).
func PlaceTreeFeature(world FeatureBlockWorld, random *WorldgenRandom, origin FeaturePosition, config TreeConfiguration) (bool, error) {
	if world == nil || random == nil || config.radius == nil || config.offset == nil {
		return false, fmt.Errorf("tree requires a world, random source and validated configuration")
	}
	tree := &treePlacement{world: world, random: random, config: config, trunks: make(map[FeaturePosition]bool), leaves: make(map[FeaturePosition]bool)}
	height := config.BaseHeight + random.NextIntBound(config.HeightRandA+1) + random.NextIntBound(config.HeightRandB+1)
	var foliageHeight int32
	if config.FoliageType == "minecraft:spruce_foliage_placer" {
		foliageHeight = max(int32(4), height-config.trunkHeight(random))
	} else {
		foliageHeight = config.foliageHeight(random)
	}
	radius := config.radius(random)
	if config.FoliageType == "minecraft:pine_foliage_placer" {
		radius += random.NextIntBound(max(height-foliageHeight+1, 1))
	}
	if origin.Y < -63 || origin.Y+height+1 > 320 {
		return false, nil
	}
	freeHeight := height
scan:
	for y := int32(0); y <= height+1; y++ {
		space := config.UpperSize
		if y < config.Limit {
			space = config.LowerSize
		}
		for x := -space; x <= space; x++ {
			for z := -space; z <= space; z++ {
				state := tree.block(treeOffset(origin, x, y, z))
				if (!treeReplaceable(state) && !treeTags["logs"][state.Name]) || (!config.IgnoreVines && state.Name == "minecraft:vine") {
					freeHeight = y - 2
					break scan
				}
			}
		}
	}
	if tree.err != nil {
		return false, tree.err
	}
	if freeHeight < height && (config.MinClippedHeight == nil || freeHeight < *config.MinClippedHeight) {
		return false, nil
	}
	below := treeOffset(origin, 0, -1, 0)
	ground := tree.block(below)
	if config.ForceDirt || !forestRockDirt[ground.Name] || ground.Name == "minecraft:grass_block" || ground.Name == "minecraft:mycelium" {
		state := config.Dirt
		if config.dirtProvider != nil {
			state = config.dirtProvider(random, below)
		}
		tree.set(below, state, tree.trunks)
		tree.trunkOrder.add(below)
	}
	var attachments []FeaturePosition
	if config.TrunkType == "minecraft:fancy_trunk_placer" {
		attachments = tree.fancyTrunk(origin, freeHeight)
	} else if config.TrunkType == "minecraft:bending_trunk_placer" {
		dir := featureHorizontalDirections[random.NextIntBound(4)]
		p := origin
		for y := int32(0); y < freeHeight; y++ {
			if y+1 >= freeHeight-1+random.NextIntBound(2) {
				p = treeOffset(p, dir.X, 0, dir.Z)
			}
			tree.log(p, "y")
			if y >= config.minLeaves {
				attachments = append(attachments, p)
			}
			p.Y++
		}
		bend := config.bendLength(random)
		for i := int32(0); i <= bend; i++ {
			tree.log(p, "y")
			attachments = append(attachments, p)
			p = treeOffset(p, dir.X, 0, dir.Z)
		}
	} else {
		for y := int32(0); y < freeHeight; y++ {
			tree.log(treeOffset(origin, 0, y, 0), "y")
		}
		attachments = []FeaturePosition{treeOffset(origin, 0, freeHeight, 0)}
	}
	for _, attachment := range attachments {
		tree.foliage(attachment, foliageHeight, radius)
	}
	if tree.err != nil {
		return false, tree.err
	}
	if len(tree.trunks) == 0 && len(tree.leaves) == 0 {
		return false, nil
	}
	tree.updateLeafDistances()
	return tree.err == nil, tree.err
}

func (tree *treePlacement) foliage(center FeaturePosition, height, radius int32) {
	offset := tree.config.offset(tree.random)
	switch tree.config.FoliageType {
	case "minecraft:random_spread_foliage_placer":
		for i := int32(0); i < tree.config.leafAttempts; i++ {
			x := tree.random.NextIntBound(radius) - tree.random.NextIntBound(radius)
			y := tree.random.NextIntBound(height) - tree.random.NextIntBound(height)
			z := tree.random.NextIntBound(radius) - tree.random.NextIntBound(radius)
			tree.leaf(treeOffset(center, x, y, z))
		}
	case "minecraft:blob_foliage_placer":
		for y := offset; y >= offset-height; y-- {
			tree.leafRow(center, max(radius-1-y/2, 0), y)
		}
	case "minecraft:fancy_foliage_placer":
		for y := offset; y >= offset-height; y-- {
			r := radius
			if y != offset && y != offset-height {
				r++
			}
			tree.leafRow(center, r, y)
		}
	case "minecraft:pine_foliage_placer":
		r := int32(0)
		for y := offset; y >= offset-height; y-- {
			tree.leafRow(center, r, y)
			if r >= 1 && y == offset-height+1 {
				r--
			} else if r < radius {
				r++
			}
		}
	case "minecraft:spruce_foliage_placer":
		r, limit, reset := tree.random.NextIntBound(2), int32(1), int32(0)
		for y := offset; y >= -height; y-- {
			tree.leafRow(center, r, y)
			if r >= limit {
				r, reset, limit = reset, 1, min(limit+1, radius)
			} else {
				r++
			}
		}
	}
}

func (tree *treePlacement) leafRow(center FeaturePosition, radius, y int32) {
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			corner := abs32(x) == radius && abs32(z) == radius
			skip := false
			switch tree.config.FoliageType {
			case "minecraft:blob_foliage_placer":
				skip = corner && (tree.random.NextIntBound(2) == 0 || y == 0)
			case "minecraft:fancy_foliage_placer":
				xf, zf := float32(abs32(x))+0.5, float32(abs32(z))+0.5
				skip = xf*xf+zf*zf > float32(radius*radius)
			default:
				skip = corner && radius > 0
			}
			if skip {
				continue
			}
			position := treeOffset(center, x, y, z)
			tree.leaf(position)
		}
	}
}

func (tree *treePlacement) leaf(p FeaturePosition) {
	old := tree.block(p)
	if !treeReplaceable(old) {
		return
	}
	state := tree.config.Foliage
	if tree.config.foliageProvider != nil {
		state = tree.config.foliageProvider(tree.random, p)
	}
	if _, ok := state.Properties["waterlogged"]; ok {
		state = stateWithProperty(state, "waterlogged", strconv.FormatBool(hasSourceWater(old)))
	}
	tree.set(p, state, tree.leaves)
}

func (tree *treePlacement) fancyLine(start, end FeaturePosition, place bool) bool {
	if !place && start == end {
		return true
	}
	delta := FeaturePosition{end.X - start.X, end.Y - start.Y, end.Z - start.Z}
	steps := max(abs32(delta.X), abs32(delta.Y), abs32(delta.Z))
	if steps == 0 {
		tree.log(start, "y")
		return true
	}
	sx, sy, sz := float32(delta.X)/float32(steps), float32(delta.Y)/float32(steps), float32(delta.Z)/float32(steps)
	for i := int32(0); i <= steps; i++ {
		position := treeOffset(start, int32(math.Floor(float64(0.5+float32(i)*sx))), int32(math.Floor(float64(0.5+float32(i)*sy))), int32(math.Floor(float64(0.5+float32(i)*sz))))
		if !place {
			state := tree.block(position)
			if !treeReplaceable(state) && !treeTags["logs"][state.Name] {
				return false
			}
			continue
		}
		axis := "y"
		x, z := abs32(position.X-start.X), abs32(position.Z-start.Z)
		if max(x, z) > 0 {
			axis = "z"
			if x >= z {
				axis = "x"
			}
		}
		tree.log(position, axis)
	}
	return tree.err == nil
}

func (tree *treePlacement) fancyTrunk(origin FeaturePosition, height int32) []FeaturePosition {
	full := height + 2
	trunkHeight := int32(math.Floor(float64(full) * 0.618))
	top := origin.Y + trunkHeight
	type branch struct {
		leaf FeaturePosition
		base int32
	}
	branches := []branch{{treeOffset(origin, 0, full-5, 0), top}}
	for y := full - 5; y >= 0; y-- {
		if float32(y) < float32(full)*0.3 {
			continue
		}
		half := float32(full) / 2
		delta := half - float32(y)
		spread := float32(math.Sqrt(float64(half*half - delta*delta)))
		if delta == 0 {
			spread = half
		} else if float32(math.Abs(float64(delta))) >= half {
			spread = 0
		}
		spread *= 0.5
		distance := float64(spread) * (float64(tree.random.NextFloat()) + 0.328)
		angle := float64(tree.random.NextFloat()*2) * math.Pi
		leaf := treeOffset(origin, int32(math.Floor(distance*math.Sin(angle)+0.5)), y-1, int32(math.Floor(distance*math.Cos(angle)+0.5)))
		if !tree.fancyLine(leaf, treeOffset(leaf, 0, 5, 0), false) {
			continue
		}
		dx, dz := origin.X-leaf.X, origin.Z-leaf.Z
		baseY := float64(leaf.Y) - math.Sqrt(float64(dx*dx+dz*dz))*0.381
		base := int32(baseY)
		if baseY > float64(top) {
			base = top
		}
		if !tree.fancyLine(FeaturePosition{origin.X, base, origin.Z}, leaf, false) {
			continue
		}
		branches = append(branches, branch{leaf, base})
	}
	tree.fancyLine(origin, treeOffset(origin, 0, trunkHeight, 0), true)
	var attachments []FeaturePosition
	for _, branch := range branches {
		base := FeaturePosition{origin.X, branch.base, origin.Z}
		if base != branch.leaf && float64(branch.base-origin.Y) >= float64(full)*0.2 {
			tree.fancyLine(base, branch.leaf, true)
		}
		if float64(branch.base-origin.Y) >= float64(full)*0.2 {
			attachments = append(attachments, branch.leaf)
		}
	}
	return attachments
}

func (tree *treePlacement) updateLeafDistances() {
	// Vanilla restricts this search to the tree's bounding box and can update
	// pre-existing leaves/logs in that box. Existing shorter distances are kept.
	lo, hi := FeaturePosition{30000000, 320, 30000000}, FeaturePosition{-30000000, -64, -30000000}
	for _, positions := range []map[FeaturePosition]bool{tree.trunks, tree.leaves} {
		for p := range positions {
			lo.X, lo.Y, lo.Z = min(lo.X, p.X), min(lo.Y, p.Y), min(lo.Z, p.Z)
			hi.X, hi.Y, hi.Z = max(hi.X, p.X), max(hi.Y, p.Y), max(hi.Z, p.Z)
		}
	}
	var pending [7]javaPositionSet
	for _, p := range tree.trunkOrder.positions() {
		pending[0].add(p)
	}
	visited := make(map[FeaturePosition]bool)
	directions := []FeaturePosition{{0, -1, 0}, {0, 1, 0}, {0, 0, -1}, {0, 0, 1}, {-1, 0, 0}, {1, 0, 0}}
	for level := 0; level < 7; {
		if pending[level].size == 0 {
			level++
			continue
		}
		position := pending[level].pop()
		visited[position] = true
		if level != 0 {
			tree.set(position, stateWithProperty(tree.block(position), "distance", strconv.Itoa(level)), tree.leaves)
		}
		for _, delta := range directions {
			next := treeOffset(position, delta.X, delta.Y, delta.Z)
			if visited[next] || next.X < lo.X || next.Y < lo.Y || next.Z < lo.Z || next.X > hi.X || next.Y > hi.Y || next.Z > hi.Z {
				continue
			}
			state := tree.block(next)
			distance, err := strconv.Atoi(state.Properties["distance"])
			if treeTags["leaves"][state.Name] && state.Properties["distance"] == "" {
				distance, err = 7, nil
			}
			if treeTags["logs"][state.Name] {
				distance, err = 0, nil
			}
			if err != nil {
				continue
			}
			distance = min(distance, level+1)
			if distance < 7 {
				pending[distance].add(next)
				level = min(level, distance)
			}
		}
	}
}
