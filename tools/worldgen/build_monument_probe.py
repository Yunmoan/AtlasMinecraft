"""Compare or publish official 1.21.1 ocean-monument reference vectors."""
import argparse
from pathlib import Path
import subprocess
import tempfile

from _bootstrap import common


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    parser.add_argument('--probe', choices=('layout', 'starts', 'blocks', 'actors'), default='layout')
    args = parser.parse_args()
    target = common.ROOT / 'artifacts/monument-reference'
    target.mkdir(parents=True, exist_ok=True)
    java, javac = common.java_tools()
    classes = target / 'classes'
    classpath = common.reference_classpath(classes)
    probes = Path(__file__).resolve().parent / 'probes'
    with tempfile.TemporaryDirectory(prefix='run-', dir=target) as temporary:
        sources = ['IglooProbeSupport.java', 'MonumentLayoutProbe.java']
        if args.probe == 'starts':
            sources += ['RuinedPortalProbeSupport.java', 'RuinedPortalStartsProbe.java', 'MonumentStartsProbe.java']
        if args.probe in ('blocks', 'actors'):
            sources += ['MonumentBlocksProbe.java']
        if args.probe == 'actors':
            sources += ['MonumentActorProbe.java']
        compiled = subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes),
                                   *(str(probes / name) for name in sources)], cwd=temporary, capture_output=True)
        common.atomic_write(target / f'{args.probe}-javac.txt', compiled.stdout + compiled.stderr)
        compiled.check_returncode()
        name = {'layout': 'MonumentLayoutProbe', 'starts': 'MonumentStartsProbe', 'blocks': 'MonumentBlocksProbe', 'actors': 'MonumentActorProbe'}[args.probe]
        result = subprocess.run([java, '-cp', classpath, f'atlasprobe.{name}'], cwd=temporary, capture_output=True)
        common.atomic_write(target / f'{args.probe}.stderr.txt', result.stderr)
        common.atomic_write(target / f'{args.probe}.stdout.txt', result.stdout)
        result.check_returncode()
    prefix, count = {'layout': (b'monument_layout ', 112), 'starts': (b'monument_start ', 504), 'blocks': (b'monument_blocks ', 192), 'actors': (b'monument_actor', 19)}[args.probe]
    rows = [line for line in result.stdout.splitlines() if line.startswith(prefix)]
    if len(rows) != count:
        raise RuntimeError(f'Expected {count} {args.probe} vectors, got {len(rows)}')
    encoded = b'\n'.join(rows) + b'\n'
    path = common.ROOT / f'testdata/worldgen/structure_monument_{args.probe}_1.21.1.txt'
    if args.check:
        if path.read_bytes() != encoded:
            raise RuntimeError(f'Original monument {args.probe} vectors changed')
    else:
        common.atomic_write(path, encoded)
    if args.probe == 'actors':
        defaults = [line.split()[1:] for line in rows if line.startswith(b'monument_actor_default ')]
        if len(defaults) != 1:
            raise RuntimeError('Expected one elder guardian default')
        encoded = b' '.join(defaults[0]) + b'\n'
        for package in ('world', 'worldgen'):
            path = common.ROOT / f'internal/{package}/monument_actor_defaults_1.21.1.txt'
            if args.check:
                if path.read_bytes() != encoded:
                    raise RuntimeError(f'Monument actor defaults changed: {path}')
            else:
                common.atomic_write(path, encoded)
    print(f'{"Checked" if args.check else "Published"} {len(rows)} original monument {args.probe} vectors')


if __name__ == '__main__':
    main()
