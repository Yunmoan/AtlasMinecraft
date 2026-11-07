"""Extract original fluid rules and every pair of native collision shapes."""
import argparse
import base64
import gzip
import json
import subprocess
from _bootstrap import common

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--check', action='store_true')
args = parser.parse_args()
target = common.ROOT / 'artifacts/fluid-rules-reference'
classes = target / 'classes'
classes.mkdir(parents=True, exist_ok=True)
java, javac = common.java_tools()
classpath = common.reference_classpath(classes)
sources = [common.ROOT / 'tools/worldgen/probes' / name for name in ('FluidRulesProbe.java', 'RuinedPortalProbeSupport.java')]
subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes), *map(str, sources)], check=True)
result = subprocess.run([java, '-Xmx768m', '-cp', classpath, 'atlasprobe.FluidRulesProbe'], cwd=target, capture_output=True)
common.atomic_write(target / 'stdout.txt', result.stdout)
common.atomic_write(target / 'stderr.txt', result.stderr)
result.check_returncode()
rows = [row[len(b'fluid_rules '):] for row in result.stdout.splitlines() if row.startswith(b'fluid_rules ')]
if len(rows) != 1:
    raise RuntimeError('Missing original fluid rule output')
data = json.loads(rows[0])
rules = base64.b64decode(data.pop('rules'), validate=True)
walls = base64.b64decode(data['walls'], validate=True)
shape_count = len(data['wall_states'])
if len(rules) != 26684 * 2 or shape_count < 320 or len(walls) != shape_count * shape_count * 6:
    raise RuntimeError('Incomplete original fluid rule coverage')
slices = {key: data.pop(key) for key in ('slice_points', 'slice_indices')}
if len(slices['slice_indices']) != 26684:
    raise RuntimeError('Incomplete original collision grids')
outputs = {
    common.ROOT / 'internal/worldgen/fluid_rules_1.21.1.bin.gz': gzip.compress(rules, mtime=0),
    common.ROOT / 'internal/worldgen/fluid_collision_slices_1.21.1.json.gz': gzip.compress(json.dumps(slices, sort_keys=True, separators=(',', ':')).encode(), mtime=0),
    common.ROOT / 'testdata/worldgen/fluid_walls_1.21.1.json.gz': gzip.compress(json.dumps(data, sort_keys=True, separators=(',', ':')).encode(), mtime=0),
}
for path, encoded in outputs.items():
    if args.check:
        if path.read_bytes() != encoded:
            raise RuntimeError(f'Original fluid rules changed: {path}')
    else:
        common.atomic_write(path, encoded)
print(f'{"Checked" if args.check else "Published"} 26,684 fluid rules, {shape_count} collision grids and {len(walls):,} native collision passages')
