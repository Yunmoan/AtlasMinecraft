"""Capture original per-section postprocessing lists, retaining order and duplicates."""
import argparse
import gzip
import subprocess
from _bootstrap import common


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    target = common.ROOT / 'artifacts/postprocessing-reference'
    classes = target / 'classes'
    java, javac = common.java_tools()
    classpath = common.reference_classpath(classes)
    sources = [common.ROOT / 'tools/worldgen/probes' / name for name in
               ['FrozenSpringProbe.java', 'TreeFeatureProbe.java', 'SurfaceChunkProbe.java']]
    subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes), *map(str, sources)], check=True)
    jobs = [('features', 'FrozenSpringProbe', ['caves_postprocessing'], b'feature_postprocessing ', 10752),
            ('terrain', 'SurfaceChunkProbe', [str(common.ROOT / 'testdata/worldgen/noise_chunk_1.21.1.txt'),
                                             'carved', str(common.reference_server()), 'postprocessing'],
             b'terrain_postprocessing ', 24)]
    for name, probe, argv, prefix, count in jobs:
        output = target / f'{name}-stdout.txt'
        with output.open('wb') as stdout, (target / f'{name}-stderr.txt').open('wb') as stderr:
            subprocess.run([java, '-Xmx1024m', '-cp', classpath, 'atlasprobe.' + probe, *argv],
                           cwd=target, stdout=stdout, stderr=stderr, check=True)
        rows = [line[len(prefix):] for line in output.read_bytes().splitlines() if line.startswith(prefix)]
        if len(rows) != count:
            raise RuntimeError(f'Incomplete original {name} capture: {len(rows)} != {count}')
        if name == 'terrain':
            # Installing the original NoiseChunk and invoking doFill must not
            # change the independently captured surface/carver block arrays.
            prefix_blocks = b'carved_chunk '
            blocks = [line for line in output.read_bytes().splitlines() if line.startswith(prefix_blocks)]
            baseline = common.ROOT / 'testdata/worldgen/carver_chunk_1.21.1.txt'
            expected = [line for line in baseline.read_bytes().splitlines() if line.startswith(prefix_blocks)]
            if blocks != expected:
                raise RuntimeError('Native doFill changed baseline carved blocks/masks')
        payload = b'\n'.join(rows) + b'\n'
        path = common.ROOT / 'testdata/worldgen' / f'{name}_postprocessing_1.21.1.jsonl.gz'
        if args.check:
            if gzip.decompress(path.read_bytes()) != payload:
                raise RuntimeError(f'Original {name} postprocessing changed')
        else:
            common.atomic_write(path, gzip.compress(payload, mtime=0))
        print(f'{"Checked" if args.check else "Published"} {len(rows)} original {name} cases', flush=True)


if __name__ == '__main__':
    main()
