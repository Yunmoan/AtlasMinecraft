from pathlib import Path
import os, subprocess

root = Path(__file__).resolve().parents[2]
target = root / 'artifacts/worldgen-probes'
cp = os.pathsep.join([str(target), str(root / 'vanilla/server-inner.jar')] +
                     [str(p) for p in (root / 'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac', '-proc:none', '-cp', cp, '-d', str(target),
                str(root / 'docs/worldgen/BiomeFilterBoundaryProbe.java')], check=True, cwd=root)
fixture = root / 'docs/worldgen/biome_filter_boundaries_1.21.1.txt'
tmp = fixture.with_suffix('.txt.tmp')
with tmp.open('w', encoding='utf-8') as output:
    subprocess.run(['java', '-cp', cp, 'atlasprobe.BiomeFilterBoundaryProbe'],
                   check=True, cwd=root, stdout=output)
rows = tmp.read_text().splitlines()
assert len(rows) == 18432 and all(row.startswith('biome_filter ') for row in rows)
tmp.replace(fixture)
