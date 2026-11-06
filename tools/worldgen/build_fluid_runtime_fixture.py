"""Sample real 1.21.1 ServerLevel fluid execution at explicit frozen tick boundaries.

All twenty-eight stone boxes are isolated from each other and the superflat terrain.
The oracle uses console commands and reads native Anvil; no Go simulation is
used to produce expectations. Commands: https://minecraft.wiki/w/Commands/tick
and https://minecraft.wiki/w/Commands/forceload.
"""
import argparse
import gzip
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
from pathlib import Path
from _bootstrap import common

CHECKPOINTS = [0, 1, 2, 3, 4, 5, 6, 7, 8, 10, 19, 20, 21, 25, 30, 60, 100]
SCENES = ['water_spread', 'lava_spread', 'water_fall', 'lava_into_water',
          'water_source_conversion', 'waterlogged_shapes', 'water_lava_contact', 'basalt',
          'campfires', 'candles', 'already_waterlogged_candles', 'slab_types',
          'double_plants', 'unsupported_plants', 'mushrooms', 'aquatic_plants',
          'leaf_chain', 'leaf_log_removed', 'wet_leaf_chain', 'leaf_stale_tick',
          'bubble_magma', 'bubble_soul_sand', 'bubble_reversed', 'bubble_removed',
          'sand_fall', 'gravel_stack', 'falling_fence', 'red_sand_water']


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    java, _ = common.java_tools()
    jar = common.reference_jar()
    go = shutil.which('go')
    if not go:
        parser.error('Go is required for read-only Anvil extraction')
    parent = common.ROOT / 'artifacts/fluid-runtime-reference'
    parent.mkdir(parents=True, exist_ok=True)
    run = Path(tempfile.mkdtemp(prefix='server-', dir=parent))
    extractor = run / 'extract-fluid-runtime.exe'
    subprocess.run([go, 'build', '-o', str(extractor), 'tools/worldgen/extract_fluid_runtime.go'],
                   cwd=common.ROOT, check=True)
    with socket.socket() as listener:
        listener.bind(('127.0.0.1', 0))
        port = listener.getsockname()[1]
    properties = dict(server_ip='127.0.0.1', server_port=port, online_mode='true',
                      level_seed=1234, level_type='minecraft:flat', generate_structures='false',
                      view_distance=2, simulation_distance=2, max_tick_time=-1, sync_chunk_writes='true')
    (run / 'eula.txt').write_text('eula=true\n')
    (run / 'server.properties').write_text(''.join(f'{k.replace("_", "-")}={v}\n' for k, v in properties.items()))
    process = subprocess.Popen([java, '-Xms256M', '-Xmx2G', '-XX:ActiveProcessorCount=4', '-jar', str(jar), 'nogui'],
                               cwd=run, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                               stderr=subprocess.STDOUT, text=True, encoding='utf-8', errors='replace')
    events, commands = queue.Queue(), []

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
                raise TimeoutError(f'Waiting for {text!r}; see {run}')
            try:
                line = events.get(timeout=remaining)
            except queue.Empty as error:
                raise TimeoutError(f'Waiting for {text!r}; see {run}') from error
            if line is None:
                raise RuntimeError(f'Original server exited before {text!r}; see {run}')
            if text in line:
                return line

    def command(text, response=None):
        commands.append(text)
        process.stdin.write(text + '\n')
        process.stdin.flush()
        if response:
            return wait_for(response)

    def extract(now):
        result = subprocess.run([str(extractor), str(run / 'world'), str(now)],
                                cwd=common.ROOT, capture_output=True)
        if result.returncode:
            raise RuntimeError(result.stderr.decode('utf-8', errors='replace'))
        return json.loads(result.stdout)

    def game_time():
        return int(re.search(r'The time is (\d+)', command('time query gametime', 'The time is '))[1])

    snapshots = []
    try:
        wait_for('Done (')
        for rule, value in [('randomTickSpeed', 0), ('doMobSpawning', 'false'), ('doFireTick', 'false'),
                            ('doWeatherCycle', 'false'), ('doDaylightCycle', 'false'), ('spawnChunkRadius', 0)]:
            command(f'gamerule {rule} {value}', f'Gamerule {rule} is now set to: {value}')
        command('tick freeze', 'The game is frozen')
        command('forceload add 16 16 79 127', 'Marked ')
        for cx in range(1, 5):
            for cz in range(1, 8):
                marker = f'ATLAS_READY_{cx}_{cz}'
                deadline = time.monotonic() + 240
                while True:
                    command(f'execute if loaded {cx*16} 64 {cz*16} run say {marker}')
                    try:
                        wait_for(marker, 2)
                        break
                    except TimeoutError:
                        if time.monotonic() > deadline:
                            raise
        for i, name in enumerate(SCENES):
            x, z = (1 + i % 4) * 16, (1 + i // 4) * 16

            def put(dx, y, dz, state):
                command(f'setblock {x+dx} {y} {z+dz} {state}', 'Changed the block')

            command(f'fill {x} 63 {z} {x+15} 70 {z+15} stone hollow', 'Successfully filled')
            if name == 'water_spread':
                put(8, 64, 8, 'water')
            elif name == 'lava_spread':
                put(8, 64, 8, 'lava')
            elif name == 'water_fall':
                put(8, 68, 8, 'water')
            elif name == 'lava_into_water':
                command(f'fill {x+1} 64 {z+1} {x+14} 64 {z+14} water', 'Successfully filled')
                put(8, 67, 8, 'lava')
            elif name == 'water_source_conversion':
                put(7, 64, 8, 'water')
                put(9, 64, 8, 'water')
            elif name == 'waterlogged_shapes':
                put(9, 64, 8, 'oak_stairs[facing=west,half=bottom,waterlogged=false]')
                put(8, 64, 9, 'oak_slab[type=bottom,waterlogged=false]')
                put(7, 64, 8, 'oak_fence[waterlogged=false]')
                put(8, 64, 7, 'cobblestone_wall[waterlogged=false]')
                put(8, 64, 8, 'water')
            elif name == 'water_lava_contact':
                put(4, 64, 8, 'lava')
                put(10, 64, 8, 'water')
            elif name == 'basalt':
                put(8, 63, 8, 'soul_soil')
                put(9, 64, 8, 'blue_ice')
                put(8, 67, 8, 'lava')
            elif name == 'campfires':
                put(9, 64, 8, 'campfire[lit=true,waterlogged=false]')
                put(7, 64, 8, 'soul_campfire[lit=true,waterlogged=false]')
                put(8, 64, 9, 'campfire[lit=false,waterlogged=false]')
                put(8, 64, 8, 'water')
            elif name in ('candles', 'already_waterlogged_candles'):
                wet = 'true' if name == 'already_waterlogged_candles' else 'false'
                for dx, dz, color, amount in [(9, 8, '', 1), (7, 8, 'red_', 2), (8, 9, 'blue_', 3), (8, 7, 'white_', 4)]:
                    put(dx, 64, dz, f'{color}candle[candles={amount},lit=true,waterlogged={wet}]')
                put(8, 64, 8, 'water')
            elif name == 'slab_types':
                put(9, 64, 8, 'stone_slab[type=double,waterlogged=false]')
                put(7, 64, 8, 'stone_slab[type=top,waterlogged=false]')
                put(8, 64, 9, 'stone_slab[type=bottom,waterlogged=false]')
                put(8, 64, 8, 'water')
            elif name == 'double_plants':
                for dx, dz, plant in [(9, 8, 'sunflower'), (9, 10, 'tall_grass'), (11, 8, 'rose_bush')]:
                    put(dx, 63, dz, 'grass_block')
                    put(dx, 64, dz, f'{plant}[half=lower]')
                    put(dx, 65, dz, f'{plant}[half=upper]')
                put(4, 64, 8, 'water')
            elif name == 'unsupported_plants':
                # Commands can place unsupported states. Later fluid shape
                # updates must remove them without replacing their own cell.
                for dx, dz, plant in [(9, 8, 'oak_sapling'), (11, 8, 'azalea'), (9, 10, 'dead_bush')]:
                    put(dx, 65, dz, plant)
                put(4, 64, 8, 'water')
            elif name == 'mushrooms':
                put(9, 64, 8, 'brown_mushroom')
                put(9, 64, 10, 'red_mushroom')
                put(4, 64, 8, 'water')
            elif name == 'aquatic_plants':
                put(9, 63, 8, 'clay')
                put(9, 64, 8, 'small_dripleaf[half=lower,waterlogged=false]')
                put(9, 65, 8, 'small_dripleaf[half=upper,waterlogged=false]')
                put(9, 64, 10, 'seagrass')
                put(4, 64, 8, 'water')
            elif name in ('leaf_chain', 'leaf_log_removed', 'wet_leaf_chain', 'leaf_stale_tick'):
                put(3, 65, 8, 'oak_log')
                for dx in range(4, 13):
                    distance = min(7, dx-3) if name == 'leaf_log_removed' else 7
                    wet = 'true' if name == 'wet_leaf_chain' else 'false'
                    leaf = 'mangrove_leaves' if dx % 2 else 'oak_leaves'
                    put(dx, 65, 8, f'{leaf}[distance={distance},persistent=true,waterlogged={wet}]')
                if name == 'leaf_log_removed':
                    put(3, 65, 8, 'air')
                elif name == 'leaf_stale_tick':
                    put(6, 65, 8, 'stone')
            elif name == 'sand_fall':
                put(8, 68, 8, 'sand')
            elif name == 'gravel_stack':
                put(8, 67, 8, 'sand')
                put(8, 68, 8, 'gravel')
            elif name == 'falling_fence':
                put(8, 64, 8, 'oak_fence')
                put(8, 68, 8, 'sand')
            elif name == 'red_sand_water':
                command(f'fill {x+1} 64 {z+1} {x+14} 65 {z+14} water', 'Successfully filled')
                put(8, 68, 8, 'red_sand')
            elif name.startswith('bubble_'):
                command(f'fill {x+7} 64 {z+7} {x+9} 69 {z+9} stone', 'Successfully filled')
                initial = 'bubble_column[drag=true]' if name in ('bubble_reversed', 'bubble_removed') else 'water'
                command(f'fill {x+8} 64 {z+8} {x+8} 69 {z+8} {initial}', 'Successfully filled')
                put(8, 63, 8, 'magma_block')
                if name in ('bubble_soul_sand', 'bubble_reversed'):
                    put(8, 63, 8, 'soul_sand')
                elif name == 'bubble_removed':
                    put(8, 63, 8, 'stone')
        start = game_time()
        previous = 0
        for tick in CHECKPOINTS:
            if tick:
                command(f'tick step {tick-previous}', 'Stepping ')
                deadline = time.monotonic() + 120
                while game_time() < start + tick:
                    if time.monotonic() > deadline:
                        raise TimeoutError('Scheduled tick step did not finish')
                    time.sleep(0.1)
                command('tick query', 'frozen')
            if game_time() != start + tick:
                raise RuntimeError('Unexpected game time outside the requested tick boundary')
            # save-all only serializes dirty native chunks. Without this
            # reversible edit, an idle queue's on-disk relative delay can be
            # from an older save even though level.dat time has advanced.
            # The marker is five blocks above the closed roof, beyond every
            # fluid in this fixture, and is restored before sampling.
            for cx in range(1, 5):
                for cz in range(1, 8):
                    command(f'setblock {cx*16} 75 {cz*16} stone', 'Changed the block')
                    command(f'setblock {cx*16} 75 {cz*16} air', 'Changed the block')
            command('save-all flush', 'Saved the game')
            snapshots.append(dict(tick=tick, chunks=extract(start + tick)))
            print(f'Original runtime checkpoint: {tick} ticks, {len(SCENES)} complete boxes', flush=True)
            previous = tick
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
        process.stdin.close()
    payload = json.dumps(dict(version=3955, min_y=60, height=16, scenes=SCENES, snapshots=snapshots), separators=(',', ':')).encode() + b'\n'
    fixture = common.ROOT / 'testdata/worldgen/fluid_runtime_1.21.1.json.gz'
    packed = gzip.compress(payload, mtime=0)
    common.atomic_write(run / fixture.name, packed)
    manifest = dict(version=common.REFERENCE_VERSION, jar_sha1=common.REFERENCE_SHA1, properties=properties,
                    commands=commands, initial_game_time=start, payload_sha256=hashlib.sha256(payload).hexdigest())
    (run / 'provenance.json').write_text(json.dumps(manifest, indent=2) + '\n')
    if args.check:
        if gzip.decompress(fixture.read_bytes()) != payload:
            raise RuntimeError(f'Fresh official runtime snapshots differ; inspect {run}')
    else:
        common.atomic_write(fixture, packed)
    print(f'{"Checked" if args.check else "Published"} original runtime snapshots; provenance: {run}', flush=True)


if __name__ == '__main__':
    main()
