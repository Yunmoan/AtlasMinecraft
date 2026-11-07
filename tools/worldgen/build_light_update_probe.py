"""Build dynamic light references with the official JAR in a temporary cwd."""
import argparse
import gzip
import os
from pathlib import Path
import subprocess
import tempfile

from _bootstrap import common

root = common.ROOT
parser = argparse.ArgumentParser()
parser.add_argument('--check', action='store_true')
args = parser.parse_args()
with tempfile.TemporaryDirectory(prefix='atlas-light-update-') as work:
    cp = common.reference_classpath(work)
    subprocess.run(['javac', '-proc:none', '-cp', cp, '-d', work,
                    str(root / 'tools/worldgen/probes/LightUpdateProbe.java')], check=True, cwd=work)
    result = subprocess.run(['java', '-cp', cp, 'atlasprobe.LightUpdateProbe'],
                            check=True, cwd=work, stdout=subprocess.PIPE)
rows = [line for line in result.stdout.decode('utf-8', errors='replace').splitlines()
        if line.startswith('light_update')]
assert sum(row.startswith('light_update_case ') for row in rows) == 9
assert sum(row.startswith('light_update_states ') for row in rows) == 81
assert sum(row.startswith('light_update_edit ') for row in rows) == 21
assert sum(row.startswith('light_update ') for row in rows) == 540
data = ('\n'.join(rows) + '\n').encode('ascii')
fixture = root / 'testdata/worldgen/light_updates_1.21.1.txt.gz'
if args.check:
    assert gzip.decompress(fixture.read_bytes()) == data, 'original light update fixture changed'
else:
    fixture.write_bytes(gzip.compress(data, mtime=0))
print(f'PASS: 9 dynamic original light cases, 21 edits, 540 complete channel arrays ({len(data)} bytes)')
