package worldgen_test

import (
	"atlas/internal/blockstate"
	"atlas/internal/worldgen"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestDimensionSurfaceAgainstVanillaJAR(t *testing.T) {
	for _,dimension:=range []string{"nether","end"} {
		t.Run(dimension,func(t *testing.T) {
			file,err:=os.Open("../../docs/worldgen/"+dimension+"_surface_1.21.1.txt");if err!=nil { t.Fatal(err) };defer file.Close()
			scan:=bufio.NewScanner(file);scan.Buffer(make([]byte,65536),4*1024*1024)
			count:=0
			ids:=make(map[string]uint32)
			for scan.Scan() {
				var label,biome string;var seed int64;var cx,cz int32
				if _,err:=fmt.Sscan(scan.Text(),&label,&seed,&cx,&cz,&biome);err!=nil || label!="surface_chunk" { t.Fatalf("invalid surface reference: %v",err) }
				compressed,err:=base64.StdEncoding.DecodeString(strings.Fields(scan.Text())[5]);if err!=nil { t.Fatal(err) }
				reader,err:=gzip.NewReader(bytes.NewReader(compressed));if err!=nil { t.Fatal(err) }
				expected,err:=io.ReadAll(io.LimitReader(reader,128*256*2+1));reader.Close()
				if err!=nil || len(expected)!=128*256*2 { t.Fatalf("invalid reference size %d (%v)",len(expected),err) }
				g,err:=worldgen.NewDimensionNoiseGenerator(seed,dimension);if err!=nil { t.Fatal(err) }
				chunk,err:=g.GenerateNoise(cx,cz);if err!=nil { t.Fatal(err) }
				biomeAt:=func(int32,int32,int32)(string,error) { return biome,nil }
				if biome=="natural" { sampler,err:=worldgen.NewDimensionBiomeSampler(seed,dimension);if err!=nil { t.Fatal(err) };biomeAt=sampler.BiomeAtQuart }
				if err:=g.ApplySurface(chunk,biomeAt);err!=nil { t.Fatal(err) }
				for i:=range chunk.Blocks {
					state,err:=chunk.BlockStateAt(int32(i%16),int32(i/256),int32(i/16%16));if err!=nil { t.Fatal(err) }
					key:=state.Name+fmt.Sprint(state.Properties)
					id,exists:=ids[key]
					if !exists { id,err=blockstate.BlockStateID(state.Name,state.Properties);if err!=nil { t.Fatal(err) };ids[key]=id }
					want:=uint32(binary.BigEndian.Uint16(expected[i*2:]))
					if id!=want { t.Fatalf("%s seed %d chunk %d,%d biome %s at %d,%d,%d got %s %v (%d) want %d",dimension,seed,cx,cz,biome,i%16,i/256,i/16%16,state.Name,state.Properties,id,want) }
				}
				count++
			}
			if err:=scan.Err();err!=nil { t.Fatal(err) };if count!=120 { t.Fatalf("surface reference count %d",count) }
		})
	}
}
