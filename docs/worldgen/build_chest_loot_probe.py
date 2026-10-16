"""Generate default chest-loot and protocol-767 menu fixtures from server.jar."""
from pathlib import Path
import argparse
import json
import os
import subprocess
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
TARGET = ROOT / 'artifacts/container-v12'

def original_assets():
    with zipfile.ZipFile(ROOT / 'vanilla/server-inner.jar') as jar:
        def read(name):
            return json.loads(jar.read('data/minecraft/' + name + '.json'))
        def tag(kind, name):
            values = []
            for entry in read('tags/' + kind + '/' + name)['values']:
                if isinstance(entry, dict):
                    entry = entry['id']
                additions = tag(kind, entry[11:]) if entry.startswith('#minecraft:') else [entry]
                for value in additions:
                    if value not in values:
                        values.append(value)
            return values
        paths = sorted(n for n in jar.namelist() if n.startswith('data/minecraft/loot_table/chests/village/') or n == 'data/minecraft/loot_table/chests/pillager_outpost.json')
        enchantments = tag('enchantment', 'on_random_loot')
        return {'tables': {'minecraft:' + n[len('data/minecraft/loot_table/'):-5]: json.loads(jar.read(n)) for n in paths},
                'enchantments': {n: {'max_level': read('enchantment/' + n[10:])['max_level']} for n in enchantments},
                'enchantment_tag': {'values': enchantments},
                'instrument_tag': {'values': tag('instrument', 'regular_goat_horns')}}

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    TARGET.mkdir(parents=True, exist_ok=True)
    cp = os.pathsep.join([str(TARGET), str(ROOT / 'vanilla/server-inner.jar')] + [str(p) for p in (ROOT / 'vanilla/libraries').rglob('*.jar')])
    compile = subprocess.run(['javac', '-proc:none', '-cp', cp, '-d', str(TARGET), *[str(HERE / n) for n in ['IglooProbeSupport.java', 'RuinedPortalProbeSupport.java', 'AtlasChestMenuPlayer.java', 'ChestLootProbe.java']]], capture_output=True, cwd=TARGET)
    (TARGET / 'chest-loot-javac.txt').write_bytes(compile.stdout+compile.stderr)
    compile.check_returncode()
    with tempfile.TemporaryDirectory(prefix='atlas-chest-loot-') as work:
        block_flags = Path(work) / 'conductor.bin'
        instruments = Path(work) / 'instruments.json'
        run = subprocess.run(['java','-Xmx512m','-cp',cp,'atlasprobe.ChestLootProbe',str(block_flags),str(ROOT/'vanilla/server-inner.jar'),str(instruments),str(TARGET)],capture_output=True,cwd=work)
        (TARGET / 'chest-loot.stdout.txt').write_bytes(run.stdout)
        (TARGET / 'chest-loot.stderr.txt').write_bytes(run.stderr)
        run.check_returncode()
        rows=[r for r in run.stdout.decode().replace('\r\n','\n').splitlines() if r.startswith(('chest_loot ', 'open_screen ', 'offhand_slot '))]
        assert len(rows)==1090, len(rows)
        menu_rows=[r for r in run.stdout.decode().replace('\r\n','\n').splitlines() if r.startswith('chest_menu ')]
        assert len(menu_rows)==1024, len(menu_rows)
        builtin_path = ROOT / 'internal/item/builtin_registries_767.json'
        builtin = json.loads(builtin_path.read_bytes())
        builtin['minecraft:instrument'] = json.loads(instruments.read_bytes())
        assets = [(HERE/'chest_loot_1.21.1.txt', ('\n'.join(rows)+'\n').encode()),
                  (HERE/'chest_menu_1.21.1.txt', ('\n'.join(menu_rows)+'\n').encode()),
                  (ROOT/'internal/world/container_conductor_1.21.1.bin',block_flags.read_bytes()),
                  (ROOT/'internal/world/container_loot_tables_1.21.1.json', json.dumps(original_assets(), separators=(',', ':')).encode()),
                  (builtin_path, json.dumps(builtin, separators=(',', ':')).encode())]
        for path, data in assets:
            if args.check:
                assert path.read_bytes()==data, path
            else:
                temp=path.with_suffix(path.suffix+'.tmp');temp.write_bytes(data);temp.replace(path)
    print('verified' if args.check else 'published', '1088 original LootTable.fill vectors and menu/redstone-conductor fixtures')

if __name__=='__main__':
    main()
