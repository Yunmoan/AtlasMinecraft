"""Verify original default campfire block-entity NBT from the village Jigsaw."""
from pathlib import Path
import argparse
import os
import subprocess
import sys
import tempfile

from _bootstrap import common

ROOT = common.ROOT
HERE = Path(__file__).resolve().parent
FIXTURES = HERE.parents[1] / "testdata/worldgen"
PROBE_DIR = HERE / "probes"
TARGET = ROOT / 'artifacts/village-worldgen-v11'
sys.path.insert(0, str(HERE))
from verify_structure_reference import atomic_write


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    TARGET.mkdir(parents=True, exist_ok=True)
    classpath = common.reference_classpath(str(TARGET))
    compiler = subprocess.run(['javac', '-proc:none', '-cp', classpath, '-d', str(TARGET),
                               *[str(PROBE_DIR / name) for name in ('IglooProbeSupport.java',
                                 'RuinedPortalProbeSupport.java', 'VillageCampfireBlockEntityProbe.java')]],
                              capture_output=True, cwd=ROOT)
    atomic_write(TARGET / 'village-campfire-javac.txt', compiler.stdout + compiler.stderr)
    compiler.check_returncode()
    with tempfile.TemporaryDirectory(prefix='atlas-village-campfire-') as work:
        process = subprocess.run(['java', '-Xmx256m', '-cp', classpath,
                                  'atlasprobe.VillageCampfireBlockEntityProbe'],
                                 capture_output=True, cwd=work)
    atomic_write(TARGET / 'village-campfire.stderr.txt', process.stderr)
    process.check_returncode()
    rows = [r for r in process.stdout.decode().replace('\r\n', '\n').splitlines()
            if r.startswith('campfire_default ')]
    assert len(rows) == 2
    fixture = FIXTURES / 'village_campfire_defaults_1.21.1.txt'
    encoded = ('\n'.join(rows) + '\n').encode()
    if args.check:
        assert fixture.read_bytes() == encoded, fixture
    else:
        atomic_write(fixture, encoded)
    print(f'{"verified" if args.check else "published"} two native campfire constructor/save defaults')


if __name__ == '__main__':
    main()
