package worldgen

import (
	"fmt"
	"math"
)

func newLegacyPerlinNoise(random *LegacyRandom, firstOctave int, amplitudes []float64) (*PerlinNoise, error) {
	if len(amplitudes) == 0 || len(amplitudes)>64 || firstOctave>0 || firstOctave+len(amplitudes)-1>0 {
		return nil, fmt.Errorf("unsupported legacy octave range %d with %d amplitudes", firstOctave,len(amplitudes))
	}
	p:=&PerlinNoise{levels:make([]*ImprovedNoise,len(amplitudes)),amplitudes:append([]float64(nil),amplitudes...),inputFactor:math.Pow(2,float64(firstOctave)),valueFactor:math.Pow(2,float64(len(amplitudes)-1))/(math.Pow(2,float64(len(amplitudes)))-1)}
	zero:=newImprovedNoise(random)
	index:=-firstOctave
	if index<len(amplitudes) && amplitudes[index]!=0 { p.levels[index]=zero }
	for i:=index-1;i>=0;i-- {
		if i<len(amplitudes) && amplitudes[i]!=0 { p.levels[i]=newImprovedNoise(random) } else {
			for j:=0;j<262;j++ { random.NextInt() }
		}
	}
	return p,nil
}

func newLegacyNormalNoise(random *LegacyRandom, firstOctave int, amplitudes []float64) (*NormalNoise,error) {
	return newNormalNoise(firstOctave,amplitudes,func() (*PerlinNoise,error) { return newLegacyPerlinNoise(random,firstOctave,amplitudes) })
}

func newModernLegacyNormalNoise(random *LegacyRandom, firstOctave int, amplitudes []float64) (*NormalNoise,error) {
	return newNormalNoise(firstOctave,amplitudes,func() (*PerlinNoise,error) {
		positional:=random.ForkPositional()
		return newPositionalPerlinNoise(firstOctave,amplitudes,func(name string) octaveRandom { return positional.FromHashOf(name) })
	})
}

// RandomState uses two sequential legacy Perlin stacks for Nether climate,
// zero shift, and modern positional octaves for other registered noise keys.
func newLegacyNamedNormalNoise(seed int64,name string) (*NormalNoise,error) {
	name=densityName(name)
	switch name {
	case "minecraft:temperature": return newLegacyNormalNoise(NewLegacyRandom(seed),-7,[]float64{1,1})
	case "minecraft:vegetation": return newLegacyNormalNoise(NewLegacyRandom(seed+1),-7,[]float64{1,1})
	case "minecraft:offset": return &NormalNoise{first:&PerlinNoise{},second:&PerlinNoise{}},nil
	}
	p,err:=Parameters(name)
	if err!=nil { return nil,err }
	random:=NewLegacyRandom(seed).ForkPositional().FromHashOf(name)
	return newModernLegacyNormalNoise(random,p.FirstOctave,p.Amplitudes)
}
