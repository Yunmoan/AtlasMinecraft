package worldgen

import "fmt"

// DimensionBiomeSampler owns its density caches; use one per generation job.
type DimensionBiomeSampler struct {
	dimension string
	settings *DensitySettings
	density *DensitySampler
	points []BiomeParameterPoint
	index *biomeSearchNode
	previous int
}

func NewDimensionBiomeSampler(seed int64,dimension string) (*DimensionBiomeSampler,error) {
	if dimension!="nether" && dimension!="end" { return nil,fmt.Errorf("unsupported dimension biome source %q",dimension) }
	settings,err:=LoadDensitySettings(dimension)
	if err!=nil { return nil,err }
	s:=&DimensionBiomeSampler{dimension:dimension,settings:settings,density:NewCellDensitySampler(seed,settings),previous:-1}
	if dimension=="nether" {
		// MultiNoiseBiomeSourceParameterList.Preset.NETHER in original order.
		for _,p:=range []struct{name string;t,h,offset float32}{
			{"minecraft:nether_wastes",0,0,0}, {"minecraft:soul_sand_valley",0,-0.5,0},
			{"minecraft:crimson_forest",0.4,0,0}, {"minecraft:warped_forest",0,0.5,0.375},
			{"minecraft:basalt_deltas",-0.5,0,0.175},
		} {
			point:=BiomeParameterPoint{Biome:p.name,Offset:QuantizeClimate(float64(p.offset))}
			point.Ranges[0]=[2]int64{QuantizeClimate(float64(p.t)),QuantizeClimate(float64(p.t))}
			point.Ranges[1]=[2]int64{QuantizeClimate(float64(p.h)),QuantizeClimate(float64(p.h))}
			s.points=append(s.points,point)
		}
		s.index=buildBiomeIndex(s.points,[]int{0,1,2,3,4})
	}
	return s,nil
}

func (s *DimensionBiomeSampler) BiomeAtQuart(qx,qy,qz int32) (string,error) {
	if s.dimension=="end" {
		cx,cz:=qx>>2,qz>>2
		if int64(cx)*int64(cx)+int64(cz)*int64(cz)<=4096 { return "minecraft:the_end",nil }
		v,err:=s.density.Sample(s.settings.Router["erosion"],(cx*2+1)*8,qy*4,(cz*2+1)*8)
		if err!=nil { return "",err }
		switch {
		case v>0.25: return "minecraft:end_highlands",nil
		case v>=-0.0625: return "minecraft:end_midlands",nil
		case v< -0.21875: return "minecraft:small_end_islands",nil
		default: return "minecraft:end_barrens",nil
		}
	}
	var climate [6]int64
	for i,field:=range climateFields {
		v,err:=s.density.Sample(s.settings.Router[field],qx*4,qy*4,qz*4)
		if err!=nil { return "",err }
		climate[i]=QuantizeClimate(v)
	}
	s.previous=s.index.nearest(s.points,climate,s.previous)
	return s.points[s.previous].Biome,nil
}
