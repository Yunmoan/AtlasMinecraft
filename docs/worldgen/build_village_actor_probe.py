"""Audit original village authored actors and capture load/finalize/save NBT."""
from pathlib import Path
import argparse
import gzip
import os
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
TARGET = ROOT / 'artifacts/worldgen-probes'
sys.path.insert(0, str(HERE))
from verify_structure_reference import atomic_write

parser = argparse.ArgumentParser()
parser.add_argument('--check', action='store_true')
parser.add_argument('--audit', action='store_true')
args = parser.parse_args()
classpath = os.pathsep.join([str(TARGET), str(ROOT/'vanilla/server-inner.jar')] + [str(p) for p in (ROOT/'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac', '-proc:none', '-cp', classpath, '-d', str(TARGET), str(HERE/'VillageActorProbe.java')], check=True, cwd=ROOT)
command = ['java', f'-Datlas.root={ROOT}', '-cp', classpath, 'atlasprobe.VillageActorProbe']
if args.audit:
    command.append('audit')
with tempfile.TemporaryDirectory(prefix='village-actor-oracle-', dir=TARGET) as work:
    process = subprocess.run(command, capture_output=True, cwd=work)
atomic_write(TARGET/'village-actors-v11.stderr.txt', process.stderr)
if process.returncode:
    errors = process.stderr.decode(errors='replace').splitlines()
    raise SystemExit('\n'.join(line for line in errors[-30:] if not line.startswith('village_actor_')))
actual = '\n'.join(line for line in process.stdout.decode(errors='replace').replace('\r\n', '\n').splitlines() if line.startswith('village_actor_')) + '\n'
atomic_write(TARGET/'village-actors-v11.stdout.txt', actual.encode())
if args.audit:
    print(process.stderr.decode().splitlines()[-1])
    raise SystemExit(0)
defaults = []
for line in actual.splitlines():
    fields = line.split()
    if fields[0] in ('village_actor_default', 'village_actor_support_default'):
        defaults.append(f'{fields[1]} {fields[3]}\n')
assert len(defaults) == 11
world_defaults = ''.join(line for line in defaults if line.split()[0] in ('minecraft:armor_stand', 'minecraft:camel'))
for path, value in ((HERE/'structure_village_actors_1.21.1.txt.gz', gzip.compress(actual.encode(), mtime=0)), (ROOT/'internal/worldgen/village_actor_defaults_1.21.1.txt', ''.join(defaults).encode()), (ROOT/'internal/world/village_actor_defaults_1.21.1.txt', world_defaults.encode())):
    if args.check:
        if path.read_bytes() != value:
            raise SystemExit('original village actor reference mismatch: '+str(path))
    else:
        atomic_write(path, value)
print(f'{"verified" if args.check else "published"} {len(actual.splitlines())} original village actor records, {sum(line.startswith("village_actor_case ") for line in actual.splitlines())} finalizations')
