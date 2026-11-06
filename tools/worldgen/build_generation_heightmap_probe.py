"""Publish/check original ProtoChunk generation-heightmap lifecycle vectors."""
import argparse
import subprocess
from _bootstrap import common

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--check', action='store_true')
args=parser.parse_args()
target=common.ROOT/'artifacts/generation-heightmap-reference'
classes=target/'classes'
classes.mkdir(parents=True,exist_ok=True)
java,javac=common.java_tools()
classpath=common.reference_classpath(classes)
source=common.ROOT/'tools/worldgen/probes/GenerationHeightmapProbe.java'
subprocess.run([javac,'-proc:none','-cp',classpath,'-d',str(classes),str(source)],check=True)
result=subprocess.run([java,'-cp',classpath,'atlasprobe.GenerationHeightmapProbe'],cwd=target,capture_output=True,check=True)
common.atomic_write(target/'stdout.txt',result.stdout)
common.atomic_write(target/'stderr.txt',result.stderr)
rows=[line for line in result.stdout.splitlines() if line.startswith(b'generation_heightmap ')]
if len(rows)!=56:
    raise RuntimeError(f'Expected 56 original heightmap rows, got {len(rows)}')
encoded=b'\n'.join(rows)+b'\n'
path=common.ROOT/'testdata/worldgen/generation_heightmaps_1.21.1.txt'
if args.check:
    if path.read_bytes()!=encoded: raise RuntimeError('Original heightmap reference changed')
else:
    common.atomic_write(path,encoded)
print(f'{"Checked" if args.check else "Published"} 56 original generation-heightmap rows')
