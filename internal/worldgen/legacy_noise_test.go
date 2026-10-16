package worldgen

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"testing"
)

func TestLegacyNoiseAgainstVanillaJAR(t *testing.T) {
	f,err:=os.Open("../../docs/worldgen/legacy_noise_1.21.1.txt");if err!=nil { t.Fatal(err) };defer f.Close()
	scan:=bufio.NewScanner(f)
	var n *NormalNoise
	var random *LegacyRandom
	var lastSeed int64
	count:=0
	for scan.Scan() {
		var kind string;var seed int64
		if _,err:=fmt.Sscan(scan.Text(),&kind,&seed);err!=nil { t.Fatal(err) }
		if n==nil || seed!=lastSeed { random=NewLegacyRandom(seed);n,err=newLegacyNormalNoise(random,-7,[]float64{1,1});if err!=nil { t.Fatal(err) };lastSeed=seed }
		var bits uint64
		if kind=="legacy_normal_next" {
			if _,err:=fmt.Sscan(scan.Text(),&kind,&seed,&bits);err!=nil { t.Fatal(err) }
			if got:=random.NextLong();got!=bits { t.Fatalf("legacy construction RNG got %d want %d",got,bits) }
		} else {
			var x,y,z float64
			if _,err:=fmt.Sscan(scan.Text(),&kind,&seed,&x,&y,&z,&bits);err!=nil { t.Fatal(err) }
			if got:=n.Noise(x,y,z);math.Float64bits(got)!=bits { t.Fatalf("%s: got %016x %g want %016x %g",scan.Text(),math.Float64bits(got),got,bits,math.Float64frombits(bits)) }
		}
		count++
	}
	if err:=scan.Err();err!=nil { t.Fatal(err) };if count!=36 { t.Fatalf("reference count %d",count) }
}
