"""Capture original bubble ticks/onPlace and generation postprocessing predicates."""
import argparse
import gzip
import subprocess
from _bootstrap import common


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    target = common.ROOT / 'artifacts/generation-block-reference'
    classes = target / 'bubble-classes'
    classes.mkdir(parents=True, exist_ok=True)
    java, javac = common.java_tools()
    classpath = common.reference_classpath(classes)
    sources = [common.ROOT / 'tools/worldgen/probes' / name
               for name in ['BlockTickLevel.java', 'BlockTickProbe.java', 'RuinedPortalProbeSupport.java']]
    subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes), *map(str, sources)], check=True)
    for mode, prefix, count, fixture in [('bubbles', b'bubble_tick ', 312, 'bubble_ticks_1.21.1.jsonl.gz'),
                                        ('postprocess', b'postprocess_states ', 7, 'generation_postprocess_1.21.1.jsonl.gz')]:
        output = target / f'{mode}-stdout.txt'
        with output.open('wb') as stdout, (target / f'{mode}-stderr.txt').open('wb') as stderr:
            subprocess.run([java, '-Xmx512m', '-cp', classpath, 'atlasprobe.BlockTickProbe', mode],
                           cwd=target, stdout=stdout, stderr=stderr, check=True)
        rows = [line[len(prefix):] for line in output.read_bytes().splitlines() if line.startswith(prefix)]
        if len(rows) != count:
            raise RuntimeError(f'Incomplete original {mode} capture: {len(rows)}')
        payload = b'\n'.join(rows) + b'\n'
        path = common.ROOT / 'testdata/worldgen' / fixture
        if args.check:
            if gzip.decompress(path.read_bytes()) != payload:
                raise RuntimeError(f'Original {mode} reference changed')
        else:
            common.atomic_write(path, gzip.compress(payload, mtime=0))
        print(f'{"Checked" if args.check else "Published"} {len(rows)} original {mode} cases', flush=True)


if __name__ == '__main__':
    main()
