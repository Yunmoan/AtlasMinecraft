from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
target = root / 'artifacts/worldgen-probes'
cp = os.pathsep.join([str(target), str(root / 'vanilla/server-inner.jar')] +
                     [str(p) for p in (root / 'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac', '-proc:none', '-cp', cp, '-d', str(target),
                str(root / 'docs/worldgen/BubbleColumnProbe.java')], check=True, cwd=root)
fixture = root / 'docs/worldgen/bubble_columns_1.21.1.txt'
tmp = fixture.with_suffix('.txt.tmp')
with tmp.open('w', encoding='utf-8') as output:
    subprocess.run(['java', '-cp', cp, 'atlasprobe.BubbleColumnProbe'],
                   check=True, cwd=root, stdout=output)
rows = tmp.read_text().splitlines()
assert len(rows) == 1008
assert sum(row.startswith('bubble_column ') for row in rows) == 960
assert sum(row.startswith('bubble_neighbor ') for row in rows) == 48
tmp.replace(fixture)
