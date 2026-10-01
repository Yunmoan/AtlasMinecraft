// Read save metadata and player positions without opening or writing the world.
package main

import (
 "atlas/internal/nbt"
 "compress/gzip"
 "fmt"
 "io"
 "os"
 "path/filepath"
)

func read(path string) nbt.Object {
 file, err:=os.Open(path); if err!=nil {panic(err)}; defer file.Close()
 gzip,err:=gzip.NewReader(file); if err!=nil {panic(err)}; defer gzip.Close()
 bytes,err:=io.ReadAll(io.LimitReader(gzip,32<<20)); if err!=nil {panic(err)}
 _,root,err:=nbt.Parse(bytes); if err!=nil {panic(err)}; result,_:=root.Object(); return result
}
func main() {
 dir:="world"; if len(os.Args)>1 {dir=os.Args[1]}
 data,_:=read(filepath.Join(dir,"level.dat")).Object("Data")
 settings,_:=data.Object("WorldGenSettings"); seed,_:=settings.Number("seed")
 marker,_:=data.Object("AtlasWorldgen"); fmt.Printf("seed=%d generator=%v\n",seed,marker)
 files,err:=filepath.Glob(filepath.Join(dir,"playerdata","*.dat")); if err!=nil {panic(err)}
 for _,file:=range files {data:=read(file); fmt.Printf("%s position=%v\n",filepath.Base(file),data["Pos"].Value)}
}
