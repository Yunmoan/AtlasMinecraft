"""Capture actual FeaturePoolElement placement with a disjoint clip and pivot."""
from pathlib import Path
import os
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
TARGET = ROOT / 'artifacts/worldgen-probes'
sys.path.insert(0, str(HERE))
from verify_structure_reference import atomic_write

classpath = os.pathsep.join([str(TARGET), str(ROOT/'vanilla/server-inner.jar')]+[str(p) for p in (ROOT/'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac','-proc:none','-cp',classpath,'-d',str(TARGET),str(HERE/'TreeFeatureProbe.java'),str(HERE/'SculkPatchProbe.java')],check=True,cwd=ROOT)
output=[]
for rotation in range(4):
    process=subprocess.run(['java','-cp',classpath,'atlasprobe.SculkPatchProbe','feature_pool',str(rotation)],capture_output=True,check=True,cwd=ROOT)
    lines=process.stdout.decode().replace('\r\n','\n').splitlines()
    output.extend(line for line in lines if line.startswith('jigsaw_feature '))
assert len(output)==480,len(output)
atomic_write(HERE/'jigsaw_feature_1.21.1.txt', ('\n'.join(output)+'\n').encode())
print(f'published {len(output)} actual FeaturePoolElement placements')
