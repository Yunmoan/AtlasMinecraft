"""Capture the original 1.21.1 chunk/level tick scheduler's observable order."""
import argparse
import gzip
import subprocess
from _bootstrap import common

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--check', action='store_true')
args = parser.parse_args()
target = common.ROOT / 'artifacts/scheduled-tick-reference'
classes = target / 'classes'
classes.mkdir(parents=True, exist_ok=True)
java, javac = common.java_tools()
classpath = common.reference_classpath(classes)
sources = [common.ROOT / 'tools/worldgen/probes' / name for name in ['ScheduledTickProbe.java', 'SavedTickNBTProbe.java']]
subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes), *map(str, sources)], check=True)
result = subprocess.run([java, '-Xmx768m', '-cp', classpath, 'atlasprobe.ScheduledTickProbe'], cwd=target, capture_output=True)
common.atomic_write(target / 'stdout.txt', result.stdout)
common.atomic_write(target / 'stderr.txt', result.stderr)
result.check_returncode()
rows = [r[len(b'scheduled_tick '):] for r in result.stdout.splitlines() if r.startswith(b'scheduled_tick ')]
if len(rows) != 168:
    raise RuntimeError(f'Incomplete original scheduler coverage: {len(rows)}')
encoded = gzip.compress(b'\n'.join(rows)+b'\n', mtime=0)
path = common.ROOT / 'testdata/worldgen/scheduled_ticks_1.21.1.jsonl.gz'
if args.check:
    if path.read_bytes() != encoded:
        raise RuntimeError('Original scheduled-tick reference changed')
else:
    common.atomic_write(path, encoded)
print(f'{"Checked" if args.check else "Published"} {len(rows)} original scheduler scenarios')
result = subprocess.run([java, '-Xmx512m', '-cp', classpath, 'atlasprobe.SavedTickNBTProbe'], cwd=target, capture_output=True)
common.atomic_write(target / 'nbt-stdout.txt', result.stdout)
common.atomic_write(target / 'nbt-stderr.txt', result.stderr)
result.check_returncode()
rows = [r[len(b'saved_tick_nbt '):] for r in result.stdout.splitlines() if r.startswith(b'saved_tick_nbt ')]
if len(rows) != 126:
    raise RuntimeError(f'Incomplete original tick NBT coverage: {len(rows)}')
encoded = gzip.compress(b'\n'.join(rows)+b'\n', mtime=0)
path = common.ROOT / 'testdata/worldgen/saved_tick_nbt_1.21.1.jsonl.gz'
if args.check:
    if path.read_bytes() != encoded:
        raise RuntimeError('Original saved-tick NBT reference changed')
else:
    common.atomic_write(path, encoded)
print(f'{"Checked" if args.check else "Published"} {len(rows)} original saved-tick NBT scenarios')

result = subprocess.run([java, '-Xmx512m', '-cp', classpath, 'atlasprobe.SavedTickNBTProbe', 'blocks'], cwd=target, capture_output=True)
common.atomic_write(target / 'block-nbt-stdout.txt', result.stdout)
common.atomic_write(target / 'block-nbt-stderr.txt', result.stderr)
result.check_returncode()
rows = [r[len(b'saved_tick_nbt '):] for r in result.stdout.splitlines() if r.startswith(b'saved_tick_nbt ')]
if len(rows) != 126:
    raise RuntimeError(f'Incomplete original block tick NBT coverage: {len(rows)}')
encoded = gzip.compress(b'\n'.join(rows)+b'\n', mtime=0)
path = common.ROOT / 'testdata/worldgen/saved_block_tick_nbt_1.21.1.jsonl.gz'
if args.check:
    if path.read_bytes() != encoded:
        raise RuntimeError('Original saved-block-tick NBT reference changed')
else:
    common.atomic_write(path, encoded)
print(f'{"Checked" if args.check else "Published"} {len(rows)} original saved-block-tick NBT scenarios')
