"""Capture all original block-state updateShape calls in six defined scenes."""
import argparse
import gzip
import subprocess
from _bootstrap import common

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--check', action='store_true')
parser.add_argument('--plants', action='store_true', help='plant subclasses with support, other-half and light contexts')
args = parser.parse_args()
target = common.ROOT / ('artifacts/plant-shape-reference' if args.plants else 'artifacts/block-shape-reference')
classes = target / 'classes'
classes.mkdir(parents=True, exist_ok=True)
java, javac = common.java_tools()
classpath = common.reference_classpath(classes)
sources = ['FluidStepLevel.java', 'FluidStepProbe.java', 'BlockShapeProbe.java', 'RuinedPortalProbeSupport.java', 'IglooProbeSupport.java']
subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes), *[str(common.ROOT / 'tools/worldgen/probes' / s) for s in sources]], check=True)
with (target / 'stdout.txt').open('wb') as stdout, (target / 'stderr.txt').open('wb') as stderr:
    subprocess.run([java, '-Xmx768m', '-cp', classpath, 'atlasprobe.BlockShapeProbe', *(['plants'] if args.plants else [])], cwd=target, stdout=stdout, stderr=stderr, check=True)
rows = []
with (target / 'stdout.txt').open('rb') as output:
    for line in output:
        if line.startswith(b'block_shape '):
            rows.append(line[len(b'block_shape '):].rstrip(b'\r\n'))
if len(rows) != (32100 if args.plants else 26684 * 36):
    raise RuntimeError(f'Incomplete native block-shape coverage: {len(rows)}')
payload = b'\n'.join(rows) + b'\n'
path = common.ROOT / ('testdata/worldgen/plant_shapes_1.21.1.jsonl.gz' if args.plants else 'testdata/worldgen/block_shapes_1.21.1.jsonl.gz')
if args.check:
    if gzip.decompress(path.read_bytes()) != payload:
        raise RuntimeError('Original block shapes changed')
else:
    common.atomic_write(path, gzip.compress(payload, mtime=0))
print(f'{"Checked" if args.check else "Published"} {len(rows)} original block-shape calls')
