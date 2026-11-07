"""Count actual terrain work with a temporary Go overlay; no runtime hooks."""
import argparse
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import os
import tempfile

from _bootstrap import common


def replace_once(source, marker, replacement):
    if source.count(marker) != 1:
        raise RuntimeError(f'Generation probe anchor changed: {marker}')
    return source.replace(marker, replacement, 1)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--workers', default='1,8')
    parser.add_argument('--radius', type=int, default=4)
    parser.add_argument('--center-x', type=int, default=40)
    parser.add_argument('--center-z', type=int, default=-40)
    parser.add_argument('--structures', action='store_true')
    parser.add_argument('--cache-limit', type=int, help='Override the terrain LRU limit only in this probe.')
    parser.add_argument('--count', type=int, default=1)
    args = parser.parse_args()
    workers = [int(value) for value in args.workers.split(',')]
    if not workers or any(w < 1 or w > 32 for w in workers) or not 0 <= args.radius <= 8 or args.count < 1:
        parser.error('Expected workers 1..32, radius 0..8 and positive count.')
    if args.cache_limit is not None and not 32 <= args.cache_limit <= 1024:
        parser.error('Expected cache limit 32..1024.')
    if not -1874998+args.radius <= args.center_x < 1874998-args.radius or not -1874998+args.radius <= args.center_z < 1874998-args.radius:
        parser.error('The requested generation neighborhood is outside world bounds.')
    go = shutil.which('go')
    if not go:
        parser.error('Go is required.')
    root = common.ROOT
    parent = root / 'artifacts/vanilla-performance/stage-probe'
    parent.mkdir(parents=True, exist_ok=True)
    output = Path(tempfile.mkdtemp(prefix='run-', dir=parent))
    temporary = output / '.tmp'
    temporary.mkdir(exist_ok=True)
    source_path = root / 'internal/worldgen/overworld.go'
    source = source_path.read_text(encoding='utf-8')
    (output / 'probe.json').write_text(json.dumps({
        'arguments': vars(args),
        'overworld_sha256': hashlib.sha256(source_path.read_bytes()).hexdigest(),
    }, indent=2), encoding='utf-8')
    if args.cache_limit is not None:
        import re
        source, matches = re.subn(r'const overworldCacheLimit = \d+',
                                  f'const overworldCacheLimit = {args.cache_limit}', source)
        if matches != 1:
            raise RuntimeError('Terrain cache limit declaration changed.')
    source = replace_once(source, 'type OverworldGenerator struct {',
                          'type OverworldGenerator struct {\n workProbe *terrainWorkProbe')
    marker = 'func (generator *OverworldGenerator) generateBase(cx, cz int32) (*overworldBase, error) {'
    source = replace_once(source, marker, marker + '''
        if p:=generator.workProbe; p!=nil {
            p.mu.Lock(); p.bases[featureChunkPosition{cx,cz}]++; p.mu.Unlock()
        }
''')
    marker = 'region.shapeRandom = NewStructureRegionRandom(generator.seed, sx, sz)'
    source = replace_once(source, marker, '''
        if p:=generator.workProbe; p!=nil {p.mu.Lock();p.sources++;p.mu.Unlock()}
''' + marker)
    source += '''
type terrainWorkProbe struct {
    mu sync.Mutex
    bases map[featureChunkPosition]int
    sources int
}
'''
    source_target = output / 'overworld.go.source'
    source_target.write_text(source, encoding='utf-8')
    benchmark = r'''package worldgen
import("fmt";"runtime";"sort";"sync";"testing")
func BenchmarkStageReuseProbe(b *testing.B){
 for _,workers:=range []int{WORKERS}{b.Run(fmt.Sprintf("workers=%d",workers),func(b *testing.B){
  var all,unique,repeated,sources,live float64
  b.ReportMetric(float64((2*RADIUS+1)*(2*RADIUS+1)),"targets/op")
  for n:=0;n<b.N;n++{
   g,err:=NewOverworldGeneratorWithOptions(1234,OverworldGeneratorOptions{GenerateStructures:STRUCTURES})
   if err!=nil{b.Fatal(err)}
   g.SetTerrainWorkers(workers)
   p:=&terrainWorkProbe{bases:make(map[featureChunkPosition]int)};g.workProbe=p
   positions:=make([]featureChunkPosition,0)
   for x:=int32(-RADIUS);x<=RADIUS;x++{for z:=int32(-RADIUS);z<=RADIUS;z++{positions=append(positions,featureChunkPosition{x,z})}}
   sort.Slice(positions,func(i,j int)bool{
    a,c:=positions[i],positions[j]; da,dc:=a.chunkX*a.chunkX+a.chunkZ*a.chunkZ,c.chunkX*c.chunkX+c.chunkZ*c.chunkZ
    if da!=dc{return da<dc};if a.chunkX!=c.chunkX{return a.chunkX<c.chunkX};return a.chunkZ<c.chunkZ
   })
   jobs:=make(chan featureChunkPosition,len(positions));for _,pos:=range positions{jobs<-pos};close(jobs)
   var wg sync.WaitGroup
   for i:=0;i<workers;i++{wg.Add(1);go func(){defer wg.Done();for pos:=range jobs{if _,err:=g.Generate(CENTERX+pos.chunkX,CENTERZ+pos.chunkZ);err!=nil{b.Error(err)}}}()}
   wg.Wait()
   if b.Failed(){b.FailNow()}
   b.StopTimer();runtime.GC();var memory runtime.MemStats;runtime.ReadMemStats(&memory);live+=float64(memory.HeapAlloc)/(1024*1024);runtime.KeepAlive(g);b.StartTimer()
   p.mu.Lock();unique+=float64(len(p.bases));sources+=float64(p.sources)
   for _,count:=range p.bases{all+=float64(count);repeated+=float64(count-1)}
   p.mu.Unlock()
  }
  b.ReportMetric(all/float64(b.N),"base_calls/op")
  b.ReportMetric(unique/float64(b.N),"unique_bases/op")
  b.ReportMetric(repeated/float64(b.N),"repeated_bases/op")
  b.ReportMetric(sources/float64(b.N),"source_visits/op")
  b.ReportMetric(live/float64(b.N),"live_heap_MiB/op")
 })}
}
'''
    for key, value in {
        'WORKERS': ','.join(map(str, workers)), 'RADIUS': str(args.radius),
        'CENTERX': str(args.center_x), 'CENTERZ': str(args.center_z),
        'STRUCTURES': str(args.structures).lower(),
    }.items():
        benchmark = benchmark.replace(key, value)
    test_target = output / 'stage_probe_test.go.source'
    test_target.write_text(benchmark, encoding='utf-8')
    overlay = output / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {
        str(source_path): str(source_target),
        str(root / 'internal/worldgen/stage_work_probe_test.go'): str(test_target),
    }}), encoding='utf-8')
    env = os.environ.copy()
    env.update(TEMP=str(temporary), TMP=str(temporary), GOTMPDIR=str(temporary))
    command = [go, 'test', '-overlay', str(overlay), './internal/worldgen', '-run', '^$',
               '-bench', '^BenchmarkStageReuseProbe$', '-benchtime=1x', '-count', str(args.count), '-timeout=15m']
    print('Probe adds counters through Go overlay; ns/op is diagnostic, not a throughput comparison.', flush=True)
    print(f'Artifacts: {output.relative_to(root)}', flush=True)
    print(f'Pinned seed=1234, structures={args.structures}, center={args.center_x},{args.center_z}, radius={args.radius}, cache_override={args.cache_limit}', flush=True)
    return subprocess.run(command, cwd=root, env=env).returncode


if __name__ == '__main__':
    raise SystemExit(main())
