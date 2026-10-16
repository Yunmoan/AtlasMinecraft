"""Capture official End City block placement and constructor/marker actors."""
from pathlib import Path
import argparse
import os
import subprocess
import sys

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
TARGET=ROOT/'artifacts/worldgen-probes'
sys.path.insert(0,str(HERE))
from verify_structure_reference import atomic_write

parser=argparse.ArgumentParser()
parser.add_argument('--check',action='store_true')
args=parser.parse_args()
classpath=os.pathsep.join([str(TARGET),str(ROOT/'vanilla/server-inner.jar')]+[str(p) for p in (ROOT/'vanilla/libraries').rglob('*.jar')])
compiler=subprocess.run(['javac','-proc:none','-cp',classpath,'-d',str(TARGET),*[str(HERE/n) for n in ('InitialAnimalSpawnProbe.java','GeneratedEntityProbe.java','GeneratedEntityPacketProbe.java','IglooProbeSupport.java','AtlasProbeGameEventDispatcher.java','EndCityBlocksProbe.java')]],capture_output=True,cwd=ROOT)
atomic_write(TARGET/'end-city-blocks-javac-v10.txt',compiler.stdout+compiler.stderr)
compiler.check_returncode()
process=subprocess.run(['java','-Xmx768m','-cp',classpath,'atlasprobe.EndCityBlocksProbe'],capture_output=True,cwd=ROOT)
atomic_write(TARGET/'end-city-blocks-v10.stderr.txt',process.stderr)
atomic_write(TARGET/'end-city-blocks-v10.stdout.txt',process.stdout)
process.check_returncode()
rows=process.stdout.decode().replace('\r\n','\n').splitlines()
blocks=[r for r in rows if r.startswith('end_city_blocks ')]
actors=[r for r in rows if r.startswith('end_city_actor_')]
defaults=[' '.join(r.split()[i] for i in (1,3)) for r in actors if r.startswith('end_city_actor_default ')]
assert len(blocks)>1000 and len(defaults)==2,(len(blocks),len(defaults))
for path,data in ((HERE/'structure_end_city_blocks_1.21.1.txt',blocks),(HERE/'structure_end_city_actors_1.21.1.txt',actors),(ROOT/'internal/world/end_city_actor_defaults_1.21.1.txt',defaults),(ROOT/'internal/worldgen/end_city_actor_defaults_1.21.1.txt',defaults)):
    value=('\n'.join(data)+'\n').encode()
    if args.check:
        assert path.read_bytes()==value,path
    else:
        atomic_write(path,value)
print(f'{"verified" if args.check else "published"} {len(blocks)} block calls and {len(actors)} actor records')
