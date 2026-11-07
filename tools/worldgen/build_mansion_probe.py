"""Publish/check independent original Java 1.21.1 woodland-mansion vectors."""
import argparse
import subprocess
import tempfile
from pathlib import Path
from _bootstrap import common

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    parser.add_argument('--probe', choices=('grid', 'layout', 'starts', 'actors', 'blocks', 'foundation', 'block_entities'), default='grid')
    parser.add_argument('--diagnose-template', help='Run a single block template and keep raw diagnostics without publishing fixtures')
    args = parser.parse_args()
    if args.diagnose_template and args.probe != 'blocks':
        parser.error('--diagnose-template requires --probe blocks')
    target = common.ROOT / 'artifacts/mansion-reference'
    target.mkdir(parents=True, exist_ok=True)
    java, javac = common.java_tools()
    classes = target / args.probe / 'classes'
    classpath = common.reference_classpath(classes)
    probes = Path(__file__).resolve().parent / 'probes'
    name = {'grid': 'MansionGridProbe', 'layout': 'MansionLayoutProbe', 'starts': 'MansionStartsProbe', 'actors': 'MansionActorProbe', 'blocks': 'MansionBlocksProbe', 'foundation': 'MansionFoundationProbe', 'block_entities': 'MansionBlockEntitiesProbe'}[args.probe]
    sources = ['IglooProbeSupport.java', name + '.java']
    if args.probe == 'starts':
        sources += ['RuinedPortalProbeSupport.java', 'RuinedPortalStartsProbe.java']
    if args.probe in ('blocks', 'block_entities'):
        sources += ['RuinedPortalProbeSupport.java']
    if args.probe in ('actors', 'blocks', 'foundation', 'block_entities'):
        sources += ['MonumentLayoutProbe.java', 'MonumentBlocksProbe.java']
    with tempfile.TemporaryDirectory(prefix='run-', dir=target) as temporary:
        compiled = subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes), *(str(probes / s) for s in sources)], cwd=temporary, capture_output=True)
        common.atomic_write(target / f'{args.probe}-javac.txt', compiled.stdout + compiled.stderr)
        compiled.check_returncode()
        label = args.probe + ('-diagnostic' if args.diagnose_template else '')
        stdout_path = target / f'{label}.stdout.txt'
        stderr_path = target / f'{label}.stderr.txt'
        extra = [args.diagnose_template] if args.diagnose_template else []
        with stdout_path.open('wb') as stdout, stderr_path.open('wb') as stderr:
            result = subprocess.run([java, '-Xmx1G', '-cp', classpath, f'atlasprobe.{name}', str(common.ROOT / 'internal/worldgen/mansion_data'), *extra], cwd=temporary, stdout=stdout, stderr=stderr)
        result.check_returncode()
        output = stdout_path.read_bytes()
    if args.diagnose_template:
        print(f'Diagnostics: {stdout_path} and {stderr_path}')
        return
    prefix = f'mansion_{"start" if args.probe == "starts" else args.probe} '.encode()
    if args.probe == 'actors':
        prefix = b'mansion_actor'
    if args.probe == 'block_entities':
        prefix = b'mansion_block_entity '
    count = {'grid': 260, 'layout': 84, 'starts': 224, 'actors': 57, 'blocks': 7008, 'foundation': 144, 'block_entities': 9}[args.probe]
    rows = [line for line in output.splitlines() if line.startswith(prefix)]
    if len(rows) != count:
        raise RuntimeError(f'Expected {count} mansion {args.probe} vectors, got {len(rows)}')
    encoded = b'\n'.join(rows) + b'\n'
    path = common.ROOT / f'testdata/worldgen/structure_mansion_{args.probe}_1.21.1.txt'
    if args.check:
        if path.read_bytes() != encoded:
            raise RuntimeError(f'Original mansion {args.probe} vectors changed')
    else:
        common.atomic_write(path, encoded)
    if args.probe == 'block_entities':
        path = common.ROOT / 'internal/worldgen/mansion_block_entities_1.21.1.txt'
        if args.check:
            if path.read_bytes() != encoded:
                raise RuntimeError('Original mansion block-entity codecs differ')
        else:
            common.atomic_write(path, encoded)
    if args.probe == 'actors':
        defaults = [b' '.join(line.split()[1:]) for line in rows if line.startswith(b'mansion_actor_default ')]
        if len(defaults) != 3:
            raise RuntimeError('Expected three mansion actor defaults')
        encoded = b'\n'.join(defaults) + b'\n'
        for package in ('world', 'worldgen'):
            path = common.ROOT / f'internal/{package}/mansion_actor_defaults_1.21.1.txt'
            if args.check:
                if path.read_bytes() != encoded:
                    raise RuntimeError(f'Original mansion defaults differ: {path}')
            else:
                common.atomic_write(path, encoded)
    print(f'{"Checked" if args.check else "Published"} {len(rows)} original mansion {args.probe} vectors')

if __name__ == '__main__':
    main()
