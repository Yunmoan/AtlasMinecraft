"""Generate fresh official worlds, then compare randomly sampled full chunks.

All worlds, arrays, logs and sample manifests are retained under --output.
No existing world or checked-in reference fixture is changed.
"""

import argparse
import concurrent.futures
from datetime import date
import json
import os
from pathlib import Path
import queue
import random
import socket
import subprocess
import threading
import time

from _bootstrap import common


def require_generation_source(head):
    """Reject reuse when the recorded generator no longer matches the checkout."""
    paths = ["internal", "go.mod"]
    changed = subprocess.run(
        ["git", "diff", "--quiet", head, "--", *paths], cwd=common.ROOT
    )
    untracked = subprocess.check_output(
        ["git", "ls-files", "--others", "--exclude-standard", "--", *paths],
        cwd=common.ROOT,
        text=True,
    )
    if changed.returncode or untracked.strip():
        raise RuntimeError(
            "Generator inputs differ from the recorded commit. Commit runtime "
            "changes and choose a fresh --output directory."
        )


def official(java, jar, directory, seed, cases, dimension="overworld"):
    if (directory / "complete.json").exists():
        return
    if (directory / "console.txt").exists():
        # Never resume ticking a partially sampled reference world: preserve
        # the old attempt and generate a fresh frozen world instead.
        parent = directory.resolve().parent
        archived = directory.with_name(directory.name + f"-incomplete-{time.time_ns()}")
        if directory.resolve().parent != parent or archived.resolve().parent != parent:
            raise ValueError(
                "Incomplete-world archive escaped the experiment directory"
            )
        directory.rename(archived)
    directory.mkdir(parents=True, exist_ok=True)
    with socket.socket() as listener:
        listener.bind(("127.0.0.1", 0))
        port = listener.getsockname()[1]
    (directory / "eula.txt").write_text("eula=true\n")
    properties = {
        "server-ip": "127.0.0.1",
        "server-port": port,
        "online-mode": "true",
        "level-seed": seed,
        "generate-structures": "true",
        "view-distance": 2,
        "simulation-distance": 2,
        "max-tick-time": -1,
        "sync-chunk-writes": "true",
        "initial-enabled-packs": "vanilla",
        "region-file-compression": "deflate",
    }
    (directory / "server.properties").write_text(
        "".join(f"{k}={v}\n" for k, v in properties.items())
    )
    process = subprocess.Popen(
        [
            java,
            "-Xms256M",
            "-Xmx3G",
            "-XX:ActiveProcessorCount=4",
            "-jar",
            str(jar),
            "nogui",
        ],
        cwd=directory,
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    events = queue.Queue()

    def consume():
        with (directory / "console.txt").open("w", encoding="utf-8") as log:
            for line in process.stdout:
                log.write(line)
                log.flush()
                events.put(line)
        events.put(None)

    reader = threading.Thread(target=consume, daemon=True)
    reader.start()

    def command(text):
        process.stdin.write(text + "\n")
        process.stdin.flush()

    def dimension_command(text):
        dimension_id = {
            "overworld": "minecraft:overworld",
            "nether": "minecraft:the_nether",
            "end": "minecraft:the_end",
        }[dimension]
        command(f"execute in {dimension_id} run {text}")

    def wait(text, timeout=600):
        deadline = time.monotonic() + timeout
        while True:
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise TimeoutError(f"{seed}: waiting for {text}; see {directory}")
            try:
                line = events.get(timeout=remaining)
            except queue.Empty:
                raise TimeoutError(f"{seed}: waiting for {text}; see {directory}")
            if line is None:
                raise RuntimeError(f"{seed}: official server exited; see {directory}")
            if text in line:
                return line

    try:
        wait("Done (")
        for rule, value in [
            ("randomTickSpeed", "0"),
            ("doMobSpawning", "false"),
            ("doWeatherCycle", "false"),
            ("doDaylightCycle", "false"),
            ("spawnChunkRadius", "0"),
        ]:
            command(f"gamerule {rule} {value}")
            wait(f"Gamerule {rule} is now set to: {value}")
        command("tick freeze")
        wait("The game is frozen")
        # Small batches keep distant chunk tickets and memory bounded.
        for start in range(0, len(cases), 10):
            batch = cases[start : start + 10]
            for c in batch:
                dimension_command(f'forceload add {c["x"]*16} {c["z"]*16}')
                wait("Marked chunk")
            for index, c in enumerate(batch, start + 1):
                x, z = c["x"] * 16, c["z"] * 16
                marker = f"ATLAS_READY_{index}"
                deadline = time.monotonic() + 600
                while True:
                    dimension_command(f"execute if loaded {x} 64 {z} run say {marker}")
                    try:
                        wait(marker, 2)
                        break
                    except TimeoutError:
                        if time.monotonic() > deadline:
                            raise
                print(
                    f'VANILLA {seed} {index}/{len(cases)} {c["x"]} {c["z"]}', flush=True
                )
            command("save-all flush")
            wait("Saved the game")
            if start + 10 < len(cases):
                dimension_command("forceload remove all")
                wait("Unmarked all force loaded chunks")
        # Keep the final tickets through shutdown so the last flushed snapshot
        # is not immediately followed by another unload transition.
        shutdown = "clean"
        command("stop")
        try:
            process.wait(timeout=60)
        except subprocess.TimeoutExpired:
            # All reference chunks were explicitly flushed above. A shutdown
            # failure must be visible, but must not discard that saved oracle.
            shutdown = "terminated_after_confirmed_save_all_flush"
            process.kill()
            process.wait()
        reader.join(timeout=10)
        if process.returncode and shutdown == "clean":
            raise RuntimeError(f"Official server exit {process.returncode}")
        common.atomic_write(
            directory / "complete.json",
            json.dumps(
                {
                    "seed": seed,
                    "dimension": dimension,
                    "chunks": len(cases),
                    "shutdown": shutdown,
                    "exit_code": process.returncode,
                }
            ).encode(),
        )
    finally:
        if process.poll() is None:
            try:
                command("stop")
                process.wait(timeout=60)
            except (OSError, subprocess.TimeoutExpired):
                process.kill()
                process.wait()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path)
    parser.add_argument(
        "--dimension", choices=["overworld", "nether", "end"], default="overworld"
    )
    parser.add_argument("--count", type=int, default=100, help="chunks per seed")
    parser.add_argument(
        "--parallel", type=int, default=2, help="simultaneous seed experiments"
    )
    parser.add_argument("--seed-limit", type=int, default=5)
    args = parser.parse_args()
    stem = "world" if args.dimension == "overworld" else args.dimension
    output = (
        args.output
        or common.ROOT
        / f"artifacts/random-{stem}-comparison-{date.today().isoformat()}"
    ).resolve()
    output.mkdir(parents=True, exist_ok=True)
    if args.count < 1 or args.parallel < 1 or not 1 <= args.seed_limit <= 5:
        parser.error(
            "count and parallel must be positive; seed-limit must be between 1 and 5"
        )
    java, _ = common.java_tools()
    jar = common.reference_jar()
    seeds = [0, 1234, -7046029254386353131, 4511451270791457643, 7122209284156038298][
        : args.seed_limit
    ]
    manifest = output / "manifest.json"
    if manifest.exists():
        config = json.loads(manifest.read_text())
        existing_dimension = (
            config["dimension"].removeprefix("minecraft:").removeprefix("the_")
        )
        if not (
            config["seeds"] == seeds
            and config["count_per_seed"] == args.count
            and existing_dimension == args.dimension
        ):
            parser.error("Existing output uses different seeds, count or dimension")
        require_generation_source(config["head"])
    else:
        head = subprocess.check_output(
            ["git", "rev-parse", "HEAD"], cwd=common.ROOT, text=True
        ).strip()
        require_generation_source(head)
        sampling_seed = 20261107 + ["overworld", "nether", "end"].index(args.dimension)
        rng = random.Random(sampling_seed)
        cases = []
        for seed in seeds:
            selected = set()
            while len(selected) < args.count:
                # Full world extent, leaving a 16-chunk halo inside the border.
                selected.add(
                    (rng.randint(-1874983, 1874982), rng.randint(-1874983, 1874982))
                )
            for x, z in sorted(selected):
                cases.append({"seed": seed, "x": x, "z": z})
        config = {
            "seeds": seeds,
            "count_per_seed": args.count,
            "sampling_seed": sampling_seed,
            "chunk_coordinate_bounds": [-1874983, 1874982],
            "dimension": args.dimension,
            "minecraft": "1.21.1",
            "structures": True,
            "frozen": True,
            "y": [-64, 319] if args.dimension == "overworld" else [0, 255],
            "head": head,
            "jar_sha1": common.REFERENCE_SHA1,
            "cases": cases,
        }
        common.atomic_write(manifest, json.dumps(config, indent=2).encode())
    binary = output / ("compare-chunks.exe" if os.name == "nt" else "compare-chunks")
    subprocess.run(
        ["go", "build", "-o", str(binary), "./tools/worldgen/compare-chunks"],
        cwd=common.ROOT,
        check=True,
    )
    validation_lock = threading.Lock()

    def experiment(seed):
        directory = output / str(seed)
        directory.mkdir(exist_ok=True)
        cases = [c for c in config["cases"] if c["seed"] == seed]
        common.atomic_write(directory / "cases.json", json.dumps(cases).encode())
        official(java, jar, directory / "vanilla", seed, cases, args.dimension)
        env = dict(os.environ, GOMAXPROCS="8")
        with (directory / "go-console.txt").open("a", encoding="utf-8") as log:
            process = subprocess.Popen(
                [
                    str(binary),
                    "-dimension",
                    args.dimension,
                    "-cases",
                    str(directory / "cases.json"),
                    "-world",
                    str(directory / "vanilla/world"),
                    "-output",
                    str(directory / "chunks"),
                    "-workers",
                    "2",
                ],
                cwd=common.ROOT,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                text=True,
                env=env,
            )
            for line in process.stdout:
                log.write(line)
                log.flush()
                print(line, end="", flush=True)
            if process.wait():
                raise RuntimeError(f"Go comparison failed for {seed}; see {directory}")
        verified = subprocess.check_output(
            [
                str(binary),
                "-dimension",
                args.dimension,
                "-verify-vanilla",
                "-cases",
                str(directory / "cases.json"),
                "-world",
                str(directory / "vanilla/world"),
                "-output",
                str(directory / "chunks"),
            ],
            cwd=common.ROOT,
            text=True,
        )
        with validation_lock:
            with (output / "final-vanilla-validation.log").open(
                "a", encoding="utf-8"
            ) as log:
                log.write(verified)
        print(verified, end="", flush=True)
        print(f"SEED_COMPLETE {seed}", flush=True)

    with concurrent.futures.ThreadPoolExecutor(max_workers=args.parallel) as pool:
        list(pool.map(experiment, seeds))
    print(f"ALL_COMPLETE {output}", flush=True)


if __name__ == "__main__":
    main()
