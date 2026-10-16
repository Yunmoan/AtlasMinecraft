"""Read bubble coordinates from the independent official-server Anvil save."""
from pathlib import Path
import argparse
import os
import shutil
import subprocess

root = Path(__file__).resolve().parents[2]
parser = argparse.ArgumentParser()
parser.add_argument('world', type=Path)
args = parser.parse_args()
go = shutil.which('go') or str(Path(os.environ['TEMP']) / 'atlas-go/go/bin/go.exe')
fixture = root / 'docs/worldgen/native_bubble_columns_1.21.1.txt'
tmp = fixture.with_suffix('.txt.tmp')
with tmp.open('w', encoding='utf-8') as output:
    subprocess.run([go, 'run', 'docs/worldgen/extract_native_bubble_columns.go', str(args.world.resolve())],
                   cwd=root, stdout=output, check=True)
rows = tmp.read_text().splitlines()
assert len(rows) == 20
assert sum(row.startswith('native_bubble_chunk ') for row in rows) == 2
assert sum(row.startswith('native_bubble_column ') for row in rows) == 18
tmp.replace(fixture)
