"""Produce independent five-village JAR block/processor/projection vectors."""
from pathlib import Path
import argparse
import os
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
TARGET = ROOT / 'artifacts/village-worldgen-v11'
sys.path.insert(0, str(HERE))
from verify_structure_reference import atomic_write


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    TARGET.mkdir(parents=True, exist_ok=True)
    classpath = os.pathsep.join([str(TARGET), str(ROOT / 'vanilla/server-inner.jar')] +
                               [str(p) for p in (ROOT / 'vanilla/libraries').rglob('*.jar')])
    sources = ('IglooProbeSupport.java', 'RuinedPortalProbeSupport.java',
               'JigsawProcessorsProbe.java', 'JigsawBlocksProbe.java',
               'JigsawVillageBlocksProbe.java')
    compiler = subprocess.run(['javac', '-proc:none', '-cp', classpath, '-d', str(TARGET),
                               *[str(HERE / name) for name in sources]],
                              capture_output=True, cwd=ROOT)
    atomic_write(TARGET / 'village-blocks-javac.txt', compiler.stdout + compiler.stderr)
    compiler.check_returncode()
    with tempfile.TemporaryDirectory(prefix='atlas-village-blocks-') as work:
        process = subprocess.run(['java', '-Xmx768m', '-cp', classpath,
                                  'atlasprobe.JigsawVillageBlocksProbe'],
                                 capture_output=True, cwd=work)
    atomic_write(TARGET / 'village-blocks.stderr.txt', process.stderr)
    atomic_write(TARGET / 'village-blocks.stdout.txt', process.stdout)
    process.check_returncode()
    rows = [r for r in process.stdout.decode().replace('\r\n', '\n').splitlines()
            if r.startswith(('jigsaw_element ', 'jigsaw_processors ', 'jigsaw_blocks '))]
    headers = [r for r in rows if r.startswith('jigsaw_element ')]
    processors = [r for r in rows if r.startswith('jigsaw_processors ')]
    placements = [r for r in rows if r.startswith('jigsaw_blocks ')]
    assert len(headers) > 400 and len(processors) == len(headers) * 32
    assert len(placements) == len(headers) * 64
    fixture = HERE / 'jigsaw_village_blocks_1.21.1.txt'
    encoded = ('\n'.join(rows) + '\n').encode()
    if args.check:
        assert fixture.read_bytes() == encoded, fixture
    else:
        atomic_write(fixture, encoded)
    print(f'{"verified" if args.check else "published"} {len(headers)} elements, '
          f'{len(processors)} processor calls, {len(placements)} clipped block calls')


if __name__ == '__main__':
    main()
