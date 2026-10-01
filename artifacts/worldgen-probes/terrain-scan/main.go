// Find reproducible mountain/plateau/cave biome positions without changing saves.
package main

import (
 "atlas/internal/worldgen"
 "fmt"
 "os"
)

func main() {
 seeds := []int64{1234, 0, 7122209284156038298, -1, 4511451270791457643}
 wanted := map[string]bool{"minecraft:jagged_peaks":true,"minecraft:frozen_peaks":true,"minecraft:stony_peaks":true,"minecraft:snowy_slopes":true,"minecraft:badlands":true,"minecraft:wooded_badlands":true,"minecraft:dripstone_caves":true,"minecraft:lush_caves":true,"minecraft:deep_dark":true}
 for _, seed := range seeds {
  sampler, err := worldgen.NewOverworldBiomeSampler(seed); if err != nil { panic(err) }
  found := map[string]bool{}
  for x:=int32(-256); x<=256 && len(found)<len(wanted); x+=8 {
   for z:=int32(-256); z<=256 && len(found)<len(wanted); z+=8 {
    for _, y:=range []int32{32,-8} {
     name, err := sampler.BiomeAtQuart(x*4,y,z*4); if err != nil { panic(err) }
     if wanted[name] && !found[name] {fmt.Printf("# %s\n%d %d %d\n", name, seed, x, z); found[name]=true}
    }
   }
  }
  if len(found)<len(wanted) { fmt.Fprintf(os.Stderr,"seed %d: found %d/%d target biomes\n",seed,len(found),len(wanted)) }
 }
}
