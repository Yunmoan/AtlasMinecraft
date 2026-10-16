"""Rebuild all official village FeaturePoolElement cases; no runtime Java."""
import argparse
import os
import subprocess
import tempfile
from pathlib import Path
from verify_structure_reference import atomic_write

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
TARGET = ROOT / 'artifacts/worldgen-probes'
parser = argparse.ArgumentParser()
parser.add_argument('--check', action='store_true')
args = parser.parse_args()
classpath = os.pathsep.join([str(TARGET), str(ROOT/'vanilla/server-inner.jar')] + [str(p) for p in (ROOT/'vanilla/libraries').rglob('*.jar')])
compiled = subprocess.run(['javac', '-proc:none', '-cp', classpath, '-d', str(TARGET), str(HERE/'TreeFeatureProbe.java'), str(HERE/'VillageFeaturePoolProbe.java')], capture_output=True, cwd=ROOT)
atomic_write(TARGET/'village-features-javac-v11.txt', compiled.stdout + compiled.stderr)
compiled.check_returncode()
with tempfile.TemporaryDirectory(prefix='atlas-village-features-') as scratch:
    process = subprocess.run(['java', '-Xmx768m', '-cp', classpath, 'atlasprobe.VillageFeaturePoolProbe'], capture_output=True, cwd=scratch)
atomic_write(TARGET/'village-features-v11.stderr.txt', process.stderr)
atomic_write(TARGET/'village-features-v11.stdout.txt', process.stdout)
process.check_returncode()
rows = [line for line in process.stdout.decode().splitlines() if line.startswith('village_feature ')]
assert len(rows) == 13*6*10*4, len(rows)
encoded = ('\n'.join(rows) + '\n').encode()
fixture = HERE/'village_feature_pool_1.21.1.txt'
if args.check:
    assert fixture.read_bytes() == encoded, fixture
else:
    atomic_write(fixture, encoded)
print(f'{"verified" if args.check else "published"} {len(rows)} actual village feature pool calls')
