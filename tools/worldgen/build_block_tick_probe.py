"""Capture original scheduled block method owners and leaf-distance ticks."""
import argparse
import gzip
import json
import subprocess
from _bootstrap import common

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--check',action='store_true')
args=parser.parse_args()
target=common.ROOT/'artifacts/block-tick-reference'
classes=target/'classes'
classes.mkdir(parents=True,exist_ok=True)
java,javac=common.java_tools()
classpath=common.reference_classpath(classes)
sources=[common.ROOT/'tools/worldgen/probes'/n for n in ['BlockTickLevel.java','BlockTickProbe.java','RuinedPortalProbeSupport.java']]
subprocess.run([javac,'-proc:none','-cp',classpath,'-d',str(classes),*map(str,sources)],check=True)
with (target/'leaf-stdout.txt').open('wb') as stdout,(target/'leaf-stderr.txt').open('wb') as stderr:
    subprocess.run([java,'-Xmx512m','-cp',classpath,'atlasprobe.BlockTickProbe'],cwd=target,stdout=stdout,stderr=stderr,check=True)
handlers,distances,rows=None,None,[]
for line in (target/'leaf-stdout.txt').read_bytes().splitlines():
    if line.startswith(b'block_tick_handlers '):handlers=json.loads(line[len(b'block_tick_handlers '):])
    elif line.startswith(b'leaf_distances '):distances=json.loads(line[len(b'leaf_distances '):])
    elif line.startswith(b'leaf_tick '):rows.append(line[len(b'leaf_tick '):])
if handlers is None or distances is None or len(distances)!=26684 or len(rows)!=14280:raise RuntimeError('Incomplete original block tick capture')
rules={owner:sorted(n for n,o in handlers.items() if o==owner) for owner in sorted(set(handlers.values()))}
payloads={
    'internal/worldgen/block_tick_handlers_1.21.1.json':(json.dumps(rules,indent=2)+'\n').encode(),
    'testdata/worldgen/leaf_distances_1.21.1.bin':bytes(distances),
    'testdata/worldgen/leaf_ticks_1.21.1.jsonl.gz':gzip.compress(b'\n'.join(rows)+b'\n',mtime=0),
}
for name,data in payloads.items():
    path=common.ROOT/name
    if args.check:
        if path.read_bytes()!=data:raise RuntimeError(f'Original block tick reference changed: {name}')
    else:common.atomic_write(path,data)
print(f'{"Checked" if args.check else "Published"} {len(handlers)} block handlers, {len(distances)} leaf distances and {len(rows)} leaf ticks')
