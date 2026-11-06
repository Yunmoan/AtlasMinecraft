"""Record original TreeFeature boundary requests and ProtoChunkTicks packing."""
import argparse
import gzip
import subprocess
from _bootstrap import common

FAMILIES = [('basic', 600), ('decorated', 840), ('giant', 600),
            ('branching', 240), ('cherry', 240), ('mangrove', 304)]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    target = common.ROOT / 'artifacts/tree-tick-reference'
    classes = target / 'classes'
    classes.mkdir(parents=True, exist_ok=True)
    java, javac = common.java_tools()
    classpath = common.reference_classpath(classes)
    source = common.ROOT / 'tools/worldgen/probes/TreeFeatureProbe.java'
    subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes), str(source)], check=True)
    rows = []
    for family, count in FAMILIES:
        output = target / f'{family}-stdout.txt'
        with output.open('wb') as stdout, (target / f'{family}-stderr.txt').open('wb') as stderr:
            subprocess.run([java, '-Xmx768m', '--add-opens', 'java.base/java.util=ALL-UNNAMED',
                            '-cp', classpath, 'atlasprobe.TreeFeatureProbe', family, 'ticks'],
                           cwd=target, stdout=stdout, stderr=stderr, check=True)
        captured = [line[len(b'tree_ticks '):] for line in output.read_bytes().splitlines()
                    if line.startswith(b'tree_ticks ')]
        if len(captured) != count:
            raise RuntimeError(f'Incomplete {family} tree tick capture: {len(captured)}')
        rows.extend(captured)
        print(f'Captured {len(captured)} original {family} tree cases', flush=True)
    payload = b'\n'.join(rows) + b'\n'
    path = common.ROOT / 'testdata/worldgen/tree_ticks_1.21.1.jsonl.gz'
    if args.check:
        if gzip.decompress(path.read_bytes()) != payload:
            raise RuntimeError('Original tree tick reference changed')
    else:
        common.atomic_write(path, gzip.compress(payload, mtime=0))
    print(f'{"Checked" if args.check else "Published"} {len(rows)} original tree tick cases')


if __name__ == '__main__':
    main()
