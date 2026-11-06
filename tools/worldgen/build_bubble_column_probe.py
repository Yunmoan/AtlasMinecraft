from pathlib import Path
import os
import subprocess

from _bootstrap import common

root = common.ROOT
target = root / 'artifacts/worldgen-probes'
java, javac = common.java_tools()
cp = common.reference_classpath(str(target))
subprocess.run([javac, '-proc:none', '-cp', cp, '-d', str(target),
                str(root / 'tools/worldgen/probes/BubbleColumnProbe.java')], check=True, cwd=root)
fixture = root / 'testdata/worldgen/bubble_columns_1.21.1.txt'
tmp = fixture.with_suffix('.txt.tmp')
with tmp.open('w', encoding='utf-8') as output:
    subprocess.run([java, '-cp', cp, 'atlasprobe.BubbleColumnProbe'],
                   check=True, cwd=root, stdout=output)
rows = tmp.read_text().splitlines()
assert len(rows) == 1008
assert sum(row.startswith('bubble_column ') for row in rows) == 960
assert sum(row.startswith('bubble_neighbor ') for row in rows) == 48
tmp.replace(fixture)
