"""Capture original placeLiquid for all container states and all five fluids."""
import argparse
import gzip
import json
import subprocess
from _bootstrap import common

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--check', action='store_true')
args = parser.parse_args()
target = common.ROOT / 'artifacts/fluid-container-reference'
classes = target / 'classes'
classes.mkdir(parents=True, exist_ok=True)
java, javac = common.java_tools()
classpath = common.reference_classpath(classes)
sources = ['FluidStepLevel.java', 'FluidStepProbe.java', 'FluidContainerProbe.java', 'RuinedPortalProbeSupport.java', 'IglooProbeSupport.java']
subprocess.run([javac, '-proc:none', '-cp', classpath, '-d', str(classes), *[str(common.ROOT / 'tools/worldgen/probes' / s) for s in sources]], check=True)
result = subprocess.run([java, '-Xmx768m', '-cp', classpath, 'atlasprobe.FluidContainerProbe'], cwd=target, capture_output=True)
common.atomic_write(target / 'stdout.txt', result.stdout)
common.atomic_write(target / 'stderr.txt', result.stderr)
result.check_returncode()
rows = [r[len(b'fluid_container '):] for r in result.stdout.splitlines() if r.startswith(b'fluid_container ')]
if len(rows) != 91970:
    raise RuntimeError(f'Incomplete original container coverage: {len(rows)}')
encoded = gzip.compress(b'\n'.join(rows)+b'\n', mtime=0)
path = common.ROOT / 'testdata/worldgen/fluid_containers_1.21.1.jsonl.gz'
sounds = [json.loads(r[len(b'container_sounds '):]) for r in result.stdout.splitlines() if r.startswith(b'container_sounds ')]
if len(sounds) != 1 or len(sounds[0]) != 2:
    raise RuntimeError('Missing original sound IDs/ranges')
sound_path = common.ROOT / 'internal/world/fluid_container_sounds_1.21.1.json'
sound_data = json.dumps(sounds[0], sort_keys=True, indent=2).encode() + b'\n'
packets = [r[len(b'container_sound_packet '):] for r in result.stdout.splitlines() if r.startswith(b'container_sound_packet ')]
if len(packets) != 2:
    raise RuntimeError('Missing native sound packet bodies')
packet_path = common.ROOT / 'testdata/worldgen/fluid_container_sound_packets_1.21.1.txt'
packet_data = b'\n'.join(packets) + b'\n'
if args.check:
    if gzip.decompress(path.read_bytes()) != gzip.decompress(encoded):
        raise RuntimeError('Original fluid container reference changed')
    if sound_path.read_bytes() != sound_data:
        raise RuntimeError('Original fluid container sounds changed')
    if packet_path.read_bytes() != packet_data:
        raise RuntimeError('Original fluid container sound packets changed')
else:
    common.atomic_write(path, encoded)
    common.atomic_write(sound_path, sound_data)
    common.atomic_write(packet_path, packet_data)
print(f'{"Checked" if args.check else "Published"} {len(rows)} original container calls')
