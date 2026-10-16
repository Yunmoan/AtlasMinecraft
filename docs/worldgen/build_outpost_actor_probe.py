"""Capture official Outpost authored actors and their STRUCTURE finalization."""
from pathlib import Path
import argparse
import os
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
TARGET = ROOT / 'artifacts/worldgen-probes'
sys.path.insert(0, str(HERE))
from verify_structure_reference import atomic_write

parser = argparse.ArgumentParser()
parser.add_argument('--check', action='store_true')
args = parser.parse_args()
classpath = os.pathsep.join([str(TARGET), str(ROOT/'vanilla/server-inner.jar')] + [str(p) for p in (ROOT/'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac', '-proc:none', '-cp', classpath, '-d', str(TARGET), *[str(HERE/name) for name in ('InitialAnimalSpawnProbe.java', 'GeneratedEntityProbe.java', 'OutpostActorProbe.java')]], check=True, cwd=ROOT)
process = subprocess.run(['java', '-cp', classpath, 'atlasprobe.OutpostActorProbe'], capture_output=True, cwd=ROOT)
atomic_write(TARGET/'outpost-actors-v8.stderr.txt', process.stderr)
if process.returncode:
    raise SystemExit(process.stderr.decode(errors='replace'))
# Setup also prints unrelated biome/creature tables by HashMap encounter order.
# Keep this fixture scoped to original Outpost records so independent JVM runs
# reproduce it without sorting any template actors or their random streams.
actual = '\n'.join(line for line in process.stdout.decode().replace('\r\n', '\n').splitlines() if line.startswith('outpost_actor_')) + '\n'
defaults = []
for line in actual.splitlines():
    fields = line.split()
    if fields[0] == 'outpost_actor_default':
        defaults.append(f'{fields[1]} {fields[3]}\n')
assert len(defaults) == 2
for path, value in ((HERE/'structure_outpost_actors_1.21.1.txt', actual), (ROOT/'internal/worldgen/outpost_actor_defaults_1.21.1.txt', ''.join(defaults)), (ROOT/'internal/world/outpost_actor_defaults_1.21.1.txt', ''.join(defaults))):
    if args.check:
        if path.read_bytes() != value.encode():
            raise SystemExit('original Outpost actor reference mismatch: '+str(path))
    else:
        atomic_write(path, value.encode())
print(f'{"verified" if args.check else "published"} {len(actual.splitlines())} original Outpost actor records')
