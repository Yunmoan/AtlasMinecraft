"""Extract/check the 73 unchanged Minecraft Java 1.21.1 mansion templates."""
import argparse
import io
import json
import zipfile
from _bootstrap import common

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    output = common.ROOT / 'internal/worldgen/mansion_data'
    with zipfile.ZipFile(common.reference_jar()) as bundle:
        version = json.loads(bundle.read('version.json'))
        if version['id'] != '1.21.1' or version['world_version'] != 3955:
            raise ValueError('expected Java 1.21.1 / DataVersion 3955')
        inner = bundle.read('META-INF/versions/1.21.1/server-1.21.1.jar')
    prefix = 'data/minecraft/structure/woodland_mansion/'
    with zipfile.ZipFile(io.BytesIO(inner)) as jar:
        expected = {n.removeprefix(prefix): jar.read(n) for n in jar.namelist() if n.startswith(prefix) and n.endswith('.nbt')}
    if len(expected) != 73:
        raise ValueError('expected exactly 73 mansion templates')
    if args.check:
        if {p.name for p in output.iterdir()} != set(expected):
            raise ValueError('mansion asset inventory differs')
        for name, data in expected.items():
            if (output / name).read_bytes() != data:
                raise ValueError('original mansion template differs: ' + name)
    else:
        output.mkdir(parents=True, exist_ok=True)
        for name, data in expected.items():
            common.atomic_write(output / name, data)
    print(f'{"Checked" if args.check else "Extracted"} 73 original mansion templates')

if __name__ == '__main__':
    main()
