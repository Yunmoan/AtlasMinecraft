package worldgen

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
	"testing"
)

func TestDimensionBiomesAgainstVanillaJAR(t *testing.T) {
	f,err:=os.Open("../../docs/worldgen/dimension_biomes_1.21.1.txt")
	if err!=nil { t.Fatal(err) }
	defer f.Close()
	scan:=bufio.NewScanner(f)
	var sampler *DimensionBiomeSampler
	var lastSeed int64
	var lastDimension string
	count:=0
	climates:=0
	for scan.Scan() {
		var kind,dimension,want string
		var seed int64
		var x,y,z int32
		if _,err:=fmt.Sscan(scan.Text(),&kind,&dimension,&seed,&x,&y,&z,&want);err!=nil { t.Fatalf("invalid dimension biome row %q: %v",scan.Text(),err) }
		if sampler==nil || seed!=lastSeed || dimension!=lastDimension {
			sampler,err=NewDimensionBiomeSampler(seed,dimension)
			if err!=nil { t.Fatal(err) }
			lastSeed,lastDimension=seed,dimension
		}
		if kind=="dimension_climate" {
			fields:=strings.Fields(scan.Text())
			for i,field:=range climateFields {
				v,err:=sampler.density.Sample(sampler.settings.Router[field],x*4,y*4,z*4)
				if err!=nil { t.Fatal(err) }
				expected,err:=strconv.ParseInt(fields[6+i],10,64)
				if err!=nil { t.Fatal(err) }
				if QuantizeClimate(v)!=expected { t.Fatalf("%s field %s got %d (%g), want %d",scan.Text(),field,QuantizeClimate(v),v,expected) }
			}
			climates++
			continue
		}
		got,err:=sampler.BiomeAtQuart(x,y,z)
		if err!=nil || got!=want { t.Fatalf("%s: got %s (%v)",scan.Text(),got,err) }
		count++
	}
	if err:=scan.Err();err!=nil { t.Fatal(err) }
	if count!=8172 { t.Fatalf("dimension biome reference count %d",count) }
	if climates!=4086 { t.Fatalf("dimension climate reference count %d",climates) }
}
