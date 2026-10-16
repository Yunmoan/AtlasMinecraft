package worldgen

import (
	"atlas/internal/blockstate"
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

type TemplateMirror uint8

const (
	TemplateMirrorNone TemplateMirror = iota
	TemplateMirrorLeftRight
	TemplateMirrorFrontBack
)

type TemplateRotation uint8

const (
	TemplateRotationNone TemplateRotation = iota
	TemplateRotationClockwise90
	TemplateRotationClockwise180
	TemplateRotationCounterclockwise90
)

//go:embed structure_transforms_1.21.1.json.gz
var templateTransformData []byte

type templateTransformTable struct {
	Version   int          `json:"version"`
	States    [][13]uint16 `json:"states"`
	Positions [][11]int32  `json:"positions"`
}

var loadTemplateTransforms = sync.OnceValues(func() (*templateTransformTable, error) {
	gz, err := gzip.NewReader(bytes.NewReader(templateTransformData))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	var result templateTransformTable
	if err := json.NewDecoder(gz).Decode(&result); err != nil {
		return nil, err
	}
	if result.Version != 3955 || len(result.States) != 26684 || len(result.Positions) != 336 {
		return nil, fmt.Errorf("incomplete original template transform table")
	}
	for id, row := range result.States {
		if row[0] > 1 || row[1] != uint16(id) {
			return nil, fmt.Errorf("invalid original template transform state %d", id)
		}
		for _, target := range row[1:] {
			if int(target) >= len(result.States) {
				return nil, fmt.Errorf("invalid original template state transform %d", target)
			}
		}
	}
	return &result, nil
})

// TransformTemplatePosition reproduces StructureTemplate.transform, including
// Java int overflow and mirror-before-rotation about the original pivot.
func TransformTemplatePosition(p, pivot FeaturePosition, mirror TemplateMirror, rotation TemplateRotation) (FeaturePosition, error) {
	if mirror > TemplateMirrorFrontBack || rotation > TemplateRotationCounterclockwise90 {
		return FeaturePosition{}, fmt.Errorf("invalid template mirror/rotation %d/%d", mirror, rotation)
	}
	x, z := p.X, p.Z
	switch mirror {
	case TemplateMirrorLeftRight:
		z = -z
	case TemplateMirrorFrontBack:
		x = -x
	}
	switch rotation {
	case TemplateRotationClockwise90:
		return FeaturePosition{pivot.X + pivot.Z - z, p.Y, pivot.Z - pivot.X + x}, nil
	case TemplateRotationClockwise180:
		return FeaturePosition{pivot.X + pivot.X - x, p.Y, pivot.Z + pivot.Z - z}, nil
	case TemplateRotationCounterclockwise90:
		return FeaturePosition{pivot.X - pivot.Z + z, p.Y, pivot.X + pivot.Z - x}, nil
	default:
		return FeaturePosition{x, p.Y, z}, nil
	}
}

// TransformTemplateBlockState uses exhaustively original-JAR-derived state
// transforms. Block subclasses' special mirrors/rotations are preserved rather
// than approximated from a generic facing property.
func TransformTemplateBlockState(state SurfaceBlockState, mirror TemplateMirror, rotation TemplateRotation) (SurfaceBlockState, error) {
	if mirror > TemplateMirrorFrontBack || rotation > TemplateRotationCounterclockwise90 {
		return SurfaceBlockState{}, fmt.Errorf("invalid template mirror/rotation %d/%d", mirror, rotation)
	}
	table, err := loadTemplateTransforms()
	if err != nil {
		return SurfaceBlockState{}, err
	}
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return SurfaceBlockState{}, err
	}
	transformed, err := blockstate.BlockStateByID(uint32(table.States[id][1+int(mirror)*4+int(rotation)]))
	if err != nil {
		return SurfaceBlockState{}, err
	}
	return SurfaceBlockState{Name: transformed.Name, Properties: transformed.Properties}, nil
}

func templateFullCube(state SurfaceBlockState) (bool, error) {
	table, err := loadTemplateTransforms()
	if err != nil {
		return false, err
	}
	id, err := blockstate.BlockStateID(state.Name, state.Properties)
	if err != nil {
		return false, err
	}
	return table.States[id][0] != 0, nil
}
