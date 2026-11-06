"""Capture original falling/brushable callbacks, entity motion and item drops."""
import argparse
import gzip
import subprocess
import hashlib
import os
import urllib.request
import zipfile
from _bootstrap import common


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    parser.add_argument('--motion', action='store_true')
    parser.add_argument('--drops', action='store_true')
    parser.add_argument('--brushable', action='store_true')
    parser.add_argument('--kelp', action='store_true', help='capture kelp scheduled survival and native Block.popResource')
    args = parser.parse_args()
    target = common.ROOT / 'artifacts/falling-block-reference'
    classes = target / 'classes'
    java, javac = common.java_tools()
    cp = common.reference_classpath(classes)
    options = []
    if args.motion and args.drops:
        parser.error('--motion and --drops are exclusive')
    if args.kelp and (args.motion or args.drops or args.brushable):
        parser.error('--kelp cannot be combined with another family or mode')
    if args.drops or args.kelp:
        asm = target / 'asm-9.8.jar'
        if not asm.is_file():
            with urllib.request.urlopen('https://repo.maven.apache.org/maven2/org/ow2/asm/asm/9.8/asm-9.8.jar', timeout=60) as response:
                common.atomic_write(asm, response.read())
        if hashlib.sha256(asm.read_bytes()).hexdigest() != '876eab6a83daecad5ca67eb9fcabb063c97b5aeb8cf1fca7a989ecde17522051':
            raise RuntimeError('Unexpected ASM dependency')
        cp += os.pathsep + str(asm)
        subprocess.run([javac, '-cp', str(asm), '-d', str(classes), str(common.ROOT / 'tools/worldgen/probes/EntityRandomProbeAgent.java')], check=True)
        agent = target / 'entity-random-agent.jar'
        with zipfile.ZipFile(agent,'w') as archive:
            archive.writestr('META-INF/MANIFEST.MF','Manifest-Version: 1.0\nPremain-Class: atlasprobe.EntityRandomProbeAgent\nClass-Path: asm-9.8.jar\n\n')
            for path in classes.rglob('EntityRandomProbeAgent*.class'):
                archive.write(path,path.relative_to(classes).as_posix())
        options = [f'-javaagent:{agent}']
    sources = [common.ROOT / 'tools/worldgen/probes' / s for s in
               ['FallingBlockLevel.java', 'FallingBlockProbe.java', 'RuinedPortalProbeSupport.java', 'IglooProbeSupport.java']]
    subprocess.run([javac, '-proc:none', '-cp', cp, '-d', str(classes), *map(str,sources)],check=True)
    mode = 'drops' if args.drops else 'motion' if args.motion else 'callbacks'
    family = 'kelp' if args.kelp else 'brushable' if args.brushable else 'falling_block'
    stem = f'{family}-{mode}' if args.brushable or args.kelp else mode
    output = target / f'{stem}-stdout.txt'
    probe_args = [mode, family] if args.brushable or args.kelp else [mode] if mode != 'callbacks' else []
    with output.open('wb') as stdout, (target / f'{stem}-stderr.txt').open('wb') as stderr:
        subprocess.run([java, '-Xmx512m', *options, '-cp', cp, 'atlasprobe.FallingBlockProbe', *probe_args],cwd=target,stdout=stdout,stderr=stderr,check=True)
    prefix = b'falling_drops ' if args.drops else b'falling_motion ' if args.motion else b'falling_block '
    rows = [s[len(prefix):] for s in output.read_bytes().splitlines() if s.startswith(prefix)]
    expected = (800 if args.drops else 992 if args.motion else 9240) if args.brushable else (1200 if args.drops else 372 if args.motion else 3360)
    if args.kelp:
        expected = 3240
    if len(rows) != expected:
        raise RuntimeError(f'Incomplete original falling block capture: {len(rows)}')
    payload = b'\n'.join(rows) + b'\n'
    fixture = common.ROOT / 'testdata/worldgen' / f'{family}_{mode}_1.21.1.jsonl.gz'
    if args.check:
        if gzip.decompress(fixture.read_bytes()) != payload:
            raise RuntimeError('Original falling block reference changed')
    else:
        common.atomic_write(fixture,gzip.compress(payload,mtime=0))
    print(f'{"Checked" if args.check else "Published"} {len(rows)} original {family} {mode} records')


if __name__ == '__main__':
    main()
