from pathlib import Path
import argparse, os, subprocess, tempfile

from _bootstrap import common

parser = argparse.ArgumentParser()
parser.add_argument('--check', action='store_true')
args = parser.parse_args()

root = common.ROOT
target = root / 'artifacts/worldgen-probes'
java, javac = common.java_tools()
cp = common.reference_classpath(str(target))
sources = ['InitialAnimalSpawnProbe', 'GeneratedEntityPacketProbe', 'AtlasProbeGameEventDispatcher', 'AtlasArmorStandPlayer', 'ArmorStandInteractionProbe']
subprocess.run([javac, '-proc:none', '-cp', cp, '-d', str(target)] + [str(root/f'tools/worldgen/probes/{s}.java') for s in sources], check=True)
fixture = root/'testdata/worldgen/armor_stand_interaction_1.21.1.txt'
with tempfile.TemporaryDirectory(prefix='atlas-armor-oracle-') as cwd:
    tmp = Path(cwd)/'armor_stand_interaction.txt'
    with tmp.open('w', encoding='utf-8') as output:
        subprocess.run([java, f'-Datlas.root={root}', '-cp', cp, 'atlasprobe.ArmorStandInteractionProbe'], check=True, cwd=cwd, stdout=output)
    result = b'\n'.join(line for line in tmp.read_bytes().splitlines() if line.startswith(b'armor_')) + b'\n'
    slots = b'\n'.join(line[len(b'armor_item_slot '):] for line in result.splitlines() if line.startswith(b'armor_item_slot ')) + b'\n'
    slot_fixture = root/'internal/world/armor_stand_item_slots_1.21.1.txt'
    if args.check:
        prior = b'\n'.join(line for line in fixture.read_bytes().splitlines() if line.startswith(b'armor_')) + b'\n'
        if prior != result or slot_fixture.read_bytes().replace(b'\r\n', b'\n') != slots:
            raise SystemExit('Armor stand original JAR vectors differ')
    else:
        fixture.write_bytes(result)
        slot_fixture.write_bytes(slots)
print(f'{"Verified" if args.check else "Created"} {fixture}')
