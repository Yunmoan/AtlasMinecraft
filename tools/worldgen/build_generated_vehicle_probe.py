from pathlib import Path
import os, subprocess

from _bootstrap import common

root = common.ROOT
target = root / 'artifacts/worldgen-probes'
cp = common.reference_classpath(str(target))
sources = ['InitialAnimalSpawnProbe', 'GeneratedEntityProbe',
           'GeneratedEntityPacketProbe', 'GeneratedVehicleProbe']
subprocess.run(['javac', '-proc:none', '-cp', cp, '-d', str(target),
                *[str(root / f'tools/worldgen/probes/{name}.java') for name in sources]],
               check=True, cwd=root)
fixture = root / 'testdata/worldgen/generated_vehicles_1.21.1.txt'
tmp = fixture.with_suffix('.txt.tmp')
with tmp.open('w', encoding='utf-8') as output:
    subprocess.run(['java', '-cp', cp, 'atlasprobe.GeneratedVehicleProbe'],
                   check=True, cwd=root, stdout=output)
tmp.replace(fixture)
defaults = [line.split() for line in fixture.read_text().splitlines() if line.startswith('vehicle_default ')]
assert len(defaults) == 1
path = root / 'internal/world/vehicle_defaults_1.21.1.txt'
tmp = path.with_suffix('.txt.tmp')
tmp.write_text(f'{defaults[0][1]} {defaults[0][3]}\n')
tmp.replace(path)
