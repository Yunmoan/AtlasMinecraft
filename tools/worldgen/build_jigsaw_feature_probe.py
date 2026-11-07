"""Capture actual FeaturePoolElement placement with a disjoint clip and pivot."""
from pathlib import Path
import os
import subprocess
import sys

from _bootstrap import common

ROOT = common.ROOT
HERE = Path(__file__).resolve().parent
FIXTURES = HERE.parents[1] / "testdata/worldgen"
PROBE_DIR = HERE / "probes"
TARGET = ROOT / 'artifacts/worldgen-probes'
sys.path.insert(0, str(HERE))
from verify_structure_reference import atomic_write

classes = TARGET / 'jigsaw-feature-classes'
classes.mkdir(parents=True, exist_ok=True)
java, javac = common.java_tools()
classpath = common.reference_classpath(classes)
subprocess.run([javac,'-proc:none','-cp',classpath,'-d',str(classes),str(PROBE_DIR /'TreeFeatureProbe.java'),str(PROBE_DIR /'SculkPatchProbe.java')],check=True,cwd=ROOT)
output=[]
for rotation in range(4):
    process=subprocess.run([java,'-cp',classpath,'atlasprobe.SculkPatchProbe','feature_pool',str(rotation)],capture_output=True,check=True,cwd=ROOT)
    lines=process.stdout.decode().replace('\r\n','\n').splitlines()
    output.extend(line for line in lines if line.startswith('jigsaw_feature '))
assert len(output)==480,len(output)
atomic_write(FIXTURES /'jigsaw_feature_1.21.1.txt', ('\n'.join(output)+'\n').encode())
print(f'published {len(output)} actual FeaturePoolElement placements')
