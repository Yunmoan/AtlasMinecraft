"""Capture actual StructureTemplate edge updates and ProtoChunkTicks packing."""
import argparse
import gzip
import subprocess
from _bootstrap import common


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    target = common.ROOT / 'artifacts/template-tick-reference'
    classes = target / 'classes'
    classes.mkdir(parents=True, exist_ok=True)
    java, javac = common.java_tools()
    cp = common.reference_classpath(classes)
    sources = [common.ROOT / 'tools/worldgen/probes' / name for name in
               ['TemplateTickLevel.java', 'TemplateTickProbe.java', 'TreeFeatureProbe.java',
                'RuinedPortalProbeSupport.java', 'IglooProbeSupport.java']]
    subprocess.run([javac, '-proc:none', '-cp', cp, '-d', str(classes), *map(str, sources)], check=True)
    output = target / 'stdout.txt'
    with output.open('wb') as stdout, (target / 'stderr.txt').open('wb') as stderr:
        subprocess.run([java, '-Xmx768m', '-cp', cp, 'atlasprobe.TemplateTickProbe'], cwd=target,
                       stdout=stdout, stderr=stderr, check=True)
    prefix = b'template_ticks '
    rows = [line[len(prefix):] for line in output.read_bytes().splitlines() if line.startswith(prefix)]
    if len(rows) != 27792:
        raise RuntimeError(f'Incomplete native template capture: {len(rows)}')
    payload = b'\n'.join(rows) + b'\n'
    fixture = common.ROOT / 'testdata/worldgen/template_ticks_1.21.1.jsonl.gz'
    if args.check:
        if gzip.decompress(fixture.read_bytes()) != payload:
            raise RuntimeError('Original template tick capture changed')
    else:
        common.atomic_write(fixture, gzip.compress(payload, mtime=0))
    print(f'{"Checked" if args.check else "Published"} {len(rows)} native template edge/finalization cases')


if __name__ == '__main__':
    main()
