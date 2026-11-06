"""Capture original LakeFeature block scheduling alongside blocks and RNG."""
import argparse
import gzip
import subprocess
from _bootstrap import common


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    target = common.ROOT / 'artifacts/generation-block-reference'
    classes = target / 'classes'
    classes.mkdir(parents=True, exist_ok=True)
    java, javac = common.java_tools()
    classpath = common.reference_classpath(classes)
    sources = [common.ROOT / 'tools/worldgen/probes' / name
               for name in ['FrozenSpringProbe.java', 'TreeFeatureProbe.java']]
    subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes), *map(str, sources)], check=True)
    output = target / 'lake-stdout.txt'
    with output.open('wb') as stdout, (target / 'lake-stderr.txt').open('wb') as stderr:
        subprocess.run([java, '-Xmx768m', '-cp', classpath, 'atlasprobe.FrozenSpringProbe', 'caves_lake_ticks'],
                       cwd=target, stdout=stdout, stderr=stderr, check=True)
    rows = [line[len(b'lake_ticks '):] for line in output.read_bytes().splitlines() if line.startswith(b'lake_ticks ')]
    if len(rows) != 512:
        raise RuntimeError(f'Incomplete original lake capture: {len(rows)}')
    payload = b'\n'.join(rows) + b'\n'
    path = common.ROOT / 'testdata/worldgen/lake_ticks_1.21.1.jsonl.gz'
    if args.check:
        if gzip.decompress(path.read_bytes()) != payload:
            raise RuntimeError('Original lake tick reference changed')
    else:
        common.atomic_write(path, gzip.compress(payload, mtime=0))
    print(f'{"Checked" if args.check else "Published"} {len(rows)} original lake cases')


if __name__ == '__main__':
    main()
