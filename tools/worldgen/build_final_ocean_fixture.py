"""Generate isolated official-server chunks and extract a final-block oracle.

Wiki command semantics: https://minecraft.wiki/w/Commands/forceload (async),
https://minecraft.wiki/w/Commands/tick. A loaded check precedes scheduled ticks.
"""
import argparse
import base64
import hashlib
import json
import queue
import re
import shutil
import socket
import subprocess
import tempfile
import threading
import time
import zipfile
import urllib.request
from pathlib import Path
from _bootstrap import common

SEED = 4511451270791457643
CHUNKS = [(0, 10), (-3, 44)]


def live_tick_fixture(trace, seed, chunks, initial_time, final_time):
    """Read-only observer records; never synthesize queues from Atlas output."""
    snapshots = {tuple(p): {} for p in chunks}
    dispatch = {tuple(p): [] for p in chunks}
    for line in trace.splitlines():
        fields = line.split()
        if fields and fields[0] == 'live_ticks':
            if len(fields) != 7:
                raise ValueError('Malformed native queue observation')
            time_value, x, z = map(int, fields[1:4])
            if (x, z) not in snapshots or time_value not in (initial_time, final_time):
                continue
            queues = fields[5:7]
            for encoded in queues:
                raw = base64.b64decode(encoded, validate=True)
                if len(raw) < 6 or raw[0] != 9:
                    raise ValueError('Native queue observation is not a ListTag')
            previous = snapshots[x, z].setdefault(time_value, queues)
            if previous != queues:
                raise ValueError('Frozen native queue changed within one time boundary')
        elif fields and fields[0] == 'scheduled_dispatch':
            if len(fields) != 7 or fields[2] not in ('block', 'fluid'):
                raise ValueError('Malformed native dispatch observation')
            time_value = int(fields[1])
            x, y, z = map(int, fields[3:6])
            key = (x >> 4, z >> 4)
            if key in dispatch and initial_time < time_value <= final_time:
                dispatch[key].append(dict(time=time_value-initial_time, kind=fields[2], position=[x, y, z], type=fields[6]))
    result = dict(version=common.REFERENCE_VERSION, seed=seed, ticks=final_time-initial_time, chunks=[])
    if result['ticks'] != 100:
        raise ValueError('Native run did not advance exactly 100 ticks')
    for x, z in chunks:
        times = snapshots[x, z]
        if initial_time not in times or final_time not in times:
            raise ValueError(f'Missing live native boundary for {x},{z}')
        if {event['kind'] for event in dispatch[x, z]} != {'block', 'fluid'}:
            raise ValueError(f'Native chunk {x},{z} did not dispatch both queues')
        result['chunks'].append(dict(x=x, z=z, initial=times[initial_time], final=times[final_time], dispatch=dispatch[x, z]))
    return (json.dumps(result, separators=(',', ':')) + '\n').encode()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true', help='compare a fresh original run with the committed fixture')
    parser.add_argument('--trace-decoration', action='store_true', help='observe original decoration calls with a Java agent')
    parser.add_argument('--preset', choices=['ocean', 'ocean-ruins', 'shipwrecks'], default='ocean')
    args = parser.parse_args()
    structures = args.preset != 'ocean'
    seed = 1234 if structures else SEED
    chunks = {'ocean': CHUNKS, 'ocean-ruins': [(-30, -56), (62, -60)],
              'shipwrecks': [(4, 12), (72, -62)]}[args.preset]
    stem = {'ocean': 'ocean', 'ocean-ruins': 'ocean_ruin', 'shipwrecks': 'shipwreck'}[args.preset]
    jar = common.reference_jar()
    java, javac = common.java_tools()
    go = shutil.which('go')
    if not go:
        parser.error('Go is required to decode original Anvil block palettes')
    parent = common.ROOT / f'artifacts/final-{args.preset}-reference'
    parent.mkdir(parents=True, exist_ok=True)
    run = Path(tempfile.mkdtemp(prefix='server-', dir=parent))
    with socket.socket() as listener:
        listener.bind(('127.0.0.1', 0))
        port = listener.getsockname()[1]
    (run / 'eula.txt').write_text('eula=true\n')
    properties = dict(server_ip='127.0.0.1', server_port=port, online_mode='true',
                      level_seed=seed, generate_structures='true' if structures else 'false', view_distance=2,
                      simulation_distance=2, max_tick_time=-1, sync_chunk_writes='true')
    (run / 'server.properties').write_text(''.join(f'{k.replace("_", "-")}={v}\n' for k, v in properties.items()))
    events = queue.Queue()
    commands = []
    java_options = []
    if args.trace_decoration:
        asm = parent / 'asm-9.8.jar'
        if not asm.is_file():
            url = 'https://repo.maven.apache.org/maven2/org/ow2/asm/asm/9.8/asm-9.8.jar'
            with urllib.request.urlopen(url, timeout=60) as response:
                common.atomic_write(asm, response.read())
        if hashlib.sha256(asm.read_bytes()).hexdigest() != '876eab6a83daecad5ca67eb9fcabb063c97b5aeb8cf1fca7a989ecde17522051':
            raise RuntimeError('Unexpected ASM 9.8 reference-tool dependency')
        classes = run / 'trace-classes'
        classes.mkdir()
        source = Path(__file__).parent / 'probes/DecorationTraceAgent.java'
        sink = source.with_name('DecorationTraceSink.java')
        subprocess.run([javac, '-cp', str(asm), '-d', str(classes), str(source), str(sink)], check=True)
        with zipfile.ZipFile(run / 'decoration-trace-sink.jar', 'w') as archive:
            path = classes / 'atlasprobe/DecorationTraceSink.class'
            archive.write(path, path.relative_to(classes).as_posix())
        agent = run / 'decoration-trace-agent.jar'
        with zipfile.ZipFile(agent, 'w') as archive:
            archive.writestr('META-INF/MANIFEST.MF', 'Manifest-Version: 1.0\nPremain-Class: atlasprobe.DecorationTraceAgent\nClass-Path: ../asm-9.8.jar\nBoot-Class-Path: decoration-trace-sink.jar\n\n')
            for path in classes.rglob('DecorationTraceAgent*.class'):
                archive.write(path, path.relative_to(classes).as_posix())
        java_options = [f'-javaagent:{agent}={run / "decoration-trace.txt"}']
    process = subprocess.Popen([java, '-Xms256M', '-Xmx2G', '-XX:ActiveProcessorCount=4', *java_options, '-jar', str(jar), 'nogui'],
                               cwd=run, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                               stderr=subprocess.STDOUT, text=True, encoding='utf-8', errors='replace')

    def consume():
        with (run / 'console.txt').open('w', encoding='utf-8') as log:
            for line in process.stdout:
                log.write(line)
                log.flush()
                events.put(line)
        events.put(None)

    reader = threading.Thread(target=consume, daemon=True)
    reader.start()

    def wait_for(text, timeout=240):
        deadline = time.monotonic() + timeout
        while True:
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise TimeoutError(f'Waiting for {text!r}; see {run / "console.txt"}')
            try:
                line = events.get(timeout=remaining)
            except queue.Empty as error:
                raise TimeoutError(f'Waiting for {text!r}; see {run / "console.txt"}') from error
            if line is None:
                raise RuntimeError(f'Official server exited before {text!r}; see {run}')
            if text in line:
                return line

    def command(text):
        commands.append(text)
        process.stdin.write(text + '\n')
        process.stdin.flush()

    def extract(world=None, mode=None):
        result = subprocess.run([go, 'run', 'tools/worldgen/extract_final_ocean_chunks.go',
                                 str(world or run / 'world'), '--preset', args.preset, *([mode] if mode else [])],
                                cwd=common.ROOT, capture_output=True)
        if result.returncode:
            raise RuntimeError(f'Original save extraction failed: {result.stderr.decode("utf-8", errors="replace")}')
        return result.stdout.replace(b'\r\n', b'\n')

    try:
        wait_for('Done (')
        for rule, value in [('randomTickSpeed', 0), ('doMobSpawning', 'false'),
                            ('doWeatherCycle', 'false'), ('doDaylightCycle', 'false'), ('spawnChunkRadius', 0)]:
            command(f'gamerule {rule} {value}')
            wait_for(f'Gamerule {rule} is now set to: {value}')
        command('tick freeze')
        wait_for('The game is frozen')
        for cx, cz in chunks:
            x, z = cx * 16, cz * 16
            command(f'forceload add {x} {z}')
            wait_for('Marked chunk')
            # forceload is asynchronous. Probe loaded status without changing blocks.
            marker = f'ATLAS_CHUNK_READY_{cx}_{cz}'
            deadline = time.monotonic() + 240
            while True:
                command(f'execute if loaded {x} 64 {z} run say {marker}')
                try:
                    wait_for(marker, 2)
                    break
                except (TimeoutError, queue.Empty):
                    if time.monotonic() > deadline:
                        raise TimeoutError(f'Chunk {cx},{cz} did not load')
            print(f'Original chunk loaded: {cx},{cz}', flush=True)
        command('save-all flush')
        wait_for('Saved the game')
        common.atomic_write(run / 'before_ticks.txt', extract())
        # Preserve pending block/fluid ticks as well as block arrays; final
        # saves overwrite those queues after they have executed.
        before_world = run / 'before-world'
        (before_world / 'region').mkdir(parents=True)
        shutil.copy2(run / 'world/level.dat', before_world / 'level.dat')
        for source in (run / 'world/region').glob('*.mca'):
            shutil.copy2(source, before_world / 'region' / source.name)
        command('time query gametime')
        initial_time = int(re.search(r'The time is (\d+)', wait_for('The time is '))[1])
        command('tick step 100')
        wait_for('Stepping 100 tick')
        deadline = time.monotonic() + 120
        while True:
            command('time query gametime')
            now = int(re.search(r'The time is (\d+)', wait_for('The time is '))[1])
            if now >= initial_time + 100:
                break
            if time.monotonic() > deadline:
                raise TimeoutError('Original scheduled-tick settling did not finish')
            time.sleep(1)
        command('tick query')
        wait_for('frozen')
        command('save-all flush')
        wait_for('Saved the game')
        command('stop')
        process.wait(timeout=120)
        reader.join(timeout=10)
        if process.returncode:
            raise RuntimeError(f'Original server failed: {process.returncode}')
    finally:
        if process.poll() is None:
            try:
                command('stop')
            except OSError:
                pass
            try:
                process.wait(timeout=30)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait()
        try:
            process.stdin.close()
        except OSError:
            pass
    rows = extract()
    fluid_rows = extract(before_world, '--fluid-ticks')
    block_rows = extract(before_world, '--block-ticks')
    entity_rows = extract(before_world, '--block-entities')
    frozen_rows = (run / 'before_ticks.txt').read_bytes()
    live_rows = None
    if args.trace_decoration:
        trace = (run / 'decoration-trace.txt').read_text()
        if 'decoration ' not in trace or ('feature_random ' not in trace and not structures):
            raise RuntimeError('Original decoration/feature hooks did not execute')
        if structures:
            live_rows = live_tick_fixture(trace, seed, chunks, initial_time, now)
    if len(rows.splitlines()) != len(chunks) or not all(row.startswith(b'final_chunk ') for row in rows.splitlines()):
        raise RuntimeError('Unexpected final-block extraction output')
    fixture = common.ROOT / f'testdata/worldgen/final_{stem}_chunk_1.21.1.txt'
    common.atomic_write(run / fixture.name, rows)
    tick_stem = f'native_{stem}_generation' if structures else 'native_generation'
    fluid_fixture = common.ROOT / f'testdata/worldgen/{tick_stem}_fluid_ticks_1.21.1.txt'
    common.atomic_write(run / fluid_fixture.name, fluid_rows)
    block_fixture = common.ROOT / f'testdata/worldgen/{tick_stem}_block_ticks_1.21.1.txt'
    frozen_fixture = common.ROOT / f'testdata/worldgen/frozen_{stem}_chunk_1.21.1.txt'
    entity_fixture = common.ROOT / f'testdata/worldgen/{tick_stem}_block_entities_1.21.1.txt'
    common.atomic_write(run / block_fixture.name, block_rows)
    live_fixture = common.ROOT / f'testdata/worldgen/native_{stem}_live_ticks_1.21.1.json'
    if live_rows is not None:
        common.atomic_write(run / live_fixture.name, live_rows)
    manifest = dict(version=common.REFERENCE_VERSION, jar_sha1=common.REFERENCE_SHA1,
                    seed=seed, chunks=chunks, properties=properties, commands=commands, trace_decoration=args.trace_decoration,
                    initial_game_time=initial_time, final_game_time=now,
                    live_ticks_sha256=hashlib.sha256(live_rows).hexdigest() if live_rows is not None else None,
                    fixture_sha256=hashlib.sha256(rows).hexdigest(),
                    fluid_fixture_sha256=hashlib.sha256(fluid_rows).hexdigest(),
                    block_fixture_sha256=hashlib.sha256(block_rows).hexdigest(),
                    frozen_fixture_sha256=hashlib.sha256(frozen_rows).hexdigest(),
                    block_entity_fixture_sha256=hashlib.sha256(entity_rows).hexdigest(),
                    source_files={str(p.relative_to(run)): hashlib.sha256(p.read_bytes()).hexdigest()
                                  for p in sorted((run / 'world').rglob('*.mca'))})
    (run / 'provenance.json').write_text(json.dumps(manifest, indent=2) + '\n')
    if args.check:
        if fixture.read_bytes() != rows:
            raise RuntimeError(f'Fresh official final chunks differ; inspect {run}')
        if fluid_fixture.read_bytes() != fluid_rows:
            raise RuntimeError(f'Fresh official generation fluid queues differ; inspect {run}')
        if block_fixture.read_bytes() != block_rows or frozen_fixture.read_bytes() != frozen_rows:
            raise RuntimeError(f'Fresh official frozen blocks/queues differ; inspect {run}')
        if structures and entity_fixture.read_bytes() != entity_rows:
            raise RuntimeError(f'Fresh official block entities differ; inspect {run}')
        if live_rows is not None and live_fixture.read_bytes() != live_rows:
            raise RuntimeError(f'Fresh official live queues/dispatch differ; inspect {run}')
    else:
        common.atomic_write(fixture, rows)
        common.atomic_write(fluid_fixture, fluid_rows)
        common.atomic_write(block_fixture, block_rows)
        common.atomic_write(frozen_fixture, frozen_rows)
        if structures:
            common.atomic_write(entity_fixture, entity_rows)
        if live_rows is not None:
            common.atomic_write(live_fixture, live_rows)
    print(f'{"Checked" if args.check else "Published"} two official final chunks; provenance: {run}', flush=True)


if __name__ == '__main__':
    main()
