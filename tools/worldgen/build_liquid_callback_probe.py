"""Capture original LiquidBlock on-place, neighbor and shape callbacks."""
import argparse
import gzip
import subprocess
from _bootstrap import common

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--check', action='store_true')
args = parser.parse_args()
target = common.ROOT / 'artifacts/liquid-callback-reference'
classes = target / 'classes'
classes.mkdir(parents=True, exist_ok=True)
java, javac = common.java_tools()
classpath = common.reference_classpath(classes)
sources = ['FluidStepLevel.java', 'FluidStepProbe.java', 'LiquidBlockCallbackProbe.java', 'RuinedPortalProbeSupport.java', 'IglooProbeSupport.java']
subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes), *[str(common.ROOT / 'tools/worldgen/probes' / s) for s in sources]], check=True)
result = subprocess.run([java, '-Xmx768m', '-cp', classpath, 'atlasprobe.LiquidBlockCallbackProbe'], cwd=target, capture_output=True)
common.atomic_write(target / 'stdout.txt', result.stdout)
common.atomic_write(target / 'stderr.txt', result.stderr)
result.check_returncode()
rows = [r[len(b'liquid_callback '):] for r in result.stdout.splitlines() if r.startswith(b'liquid_callback ')]
if len(rows) != 4992:
    raise RuntimeError(f'Incomplete original liquid callback coverage: {len(rows)}')
encoded = gzip.compress(b'\n'.join(rows)+b'\n', mtime=0)
path = common.ROOT / 'testdata/worldgen/liquid_callbacks_1.21.1.jsonl.gz'
if args.check:
    if path.read_bytes() != encoded:
        raise RuntimeError('Original liquid callback reference changed')
else:
    common.atomic_write(path, encoded)
print(f'{"Checked" if args.check else "Published"} {len(rows)} original liquid callbacks')
