"""Compare upstream and current chunk pipelines with one identical benchmark.

Requires a clean internal/ and go.mod. Baseline Go files are supplied through
an overlay; the current checkout and runtime saves are never replaced.
"""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import statistics
import subprocess

ROOT = Path(__file__).resolve().parents[1]
HARNESS = "internal/world/chunk_pipeline_performance_test.go"
PATTERN = "Benchmark(GenerateNormalArea|LoadSavedNormalArea)/workers=8$"
ROW = re.compile(
    r"^Benchmark(GenerateNormalArea|LoadSavedNormalArea)/workers=8-\d+\s+"
    r"\d+\s+(\d+) ns/op\s+([\d.]+) chunks/op\s+(\d+) B/op\s+(\d+) allocs/op$",
    re.MULTILINE,
)


def git(*args):
    return subprocess.check_output(["git", *args], cwd=ROOT)


def require_clean():
    if git("status", "--porcelain", "--", "internal", "go.mod").strip():
        raise RuntimeError("Commit runtime changes before benchmarking.")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--baseline", required=True, help="Original Git commit/ref.")
    parser.add_argument("--output", type=Path, required=True, help="Fresh evidence directory.")
    parser.add_argument("--pairs", type=int, default=3)
    args = parser.parse_args()
    if not 1 <= args.pairs <= 10:
        parser.error("pairs must be between 1 and 10")
    go = shutil.which("go")
    if not go:
        parser.error("Go is required")
    require_clean()
    baseline = git("rev-parse", args.baseline + "^{commit}").decode().strip()
    candidate = git("rev-parse", "HEAD").decode().strip()
    changed = git("diff", "--name-only", baseline, candidate, "--", "internal", "go.mod").decode().splitlines()
    if any(not path.endswith(".go") for path in changed):
        raise RuntimeError("This overlay runner requires unchanged embedded inputs and go.mod.")
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=False)
    temporary = output / ".tmp"
    temporary.mkdir()
    snapshots = output / "overlay-sources"
    snapshots.mkdir()
    old_paths = set(git("ls-tree", "-r", "--name-only", baseline, "--", "internal").decode().splitlines())
    replacements = {}
    for path in changed:
        if path == HARNESS:
            continue
        source = ROOT / path
        if path in old_paths:
            data = git("show", f"{baseline}:{path}")
        else:
            package = re.search(rb"^package (\w+)", source.read_bytes(), re.MULTILINE)
            if not package:
                raise RuntimeError(f"Cannot determine package for {path}")
            data = b"package " + package[1] + b"\n"
        target = snapshots / (path.replace("/", "_") + ".source")
        target.write_bytes(data)
        replacements[str(source)] = str(target)
    overlay = output / "overlay.json"
    overlay.write_text(json.dumps({"Replace": replacements}, indent=2), encoding="utf-8")
    environment = dict(os.environ, TEMP=str(temporary), TMP=str(temporary),
                       TMPDIR=str(temporary), GOTMPDIR=str(temporary), GOMAXPROCS="16")
    metadata = {
        "baseline": baseline,
        "candidate": candidate,
        "baseline_internal_tree": git("rev-parse", f"{baseline}:internal").decode().strip(),
        "candidate_internal_tree": git("rev-parse", f"{candidate}:internal").decode().strip(),
        "go": subprocess.check_output([go, "version"], text=True).strip(),
        "go_environment": json.loads(subprocess.check_output([go, "env", "-json", "GOOS", "GOARCH", "GOAMD64"], text=True)),
        "gomaxprocs": 16, "workers": 8, "pairs": args.pairs,
        "seed": 1234, "beehive_salt": 0, "structures": True,
        "center": [40, -40], "radius": 4, "chunks_per_operation": 81,
        "harness": HARNESS,
        "harness_sha256": hashlib.sha256((ROOT / HARNESS).read_bytes()).hexdigest(),
        "baseline_overlay": sorted(path for path in changed if path != HARNESS),
        "timing": "Production generation/lighting/save or saved-area load; excludes Create/Open/Close. Each sample uses a fresh process. Saved-area loading has warm OS file caches.",
        "host_controls": "Desktop background load, CPU frequency and power state are not locked.",
    }
    (output / "environment.json").write_text(json.dumps(metadata, indent=2) + "\n", encoding="utf-8")
    binaries = {}
    for label in ("before", "after"):
        binary = output / (label + (".test.exe" if os.name == "nt" else ".test"))
        command = [go, "test", "-c"]
        if label == "before":
            command += ["-overlay", str(overlay)]
        command += ["./internal/world", "-o", str(binary)]
        print(f"BUILD {label}", flush=True)
        subprocess.run(command, cwd=ROOT, env=environment, check=True)
        binaries[label] = binary
    require_clean()
    trials = []
    for pair in range(1, args.pairs + 1):
        order = ("before", "after") if pair % 2 else ("after", "before")
        for label in order:
            command = [str(binaries[label]), "-test.run=^$", f"-test.bench={PATTERN}",
                       "-test.benchtime=1x", "-test.count=1", "-test.timeout=15m"]
            path = output / f"pair-{pair}-{label}.txt"
            print(f"RUN pair={pair} {label}", flush=True)
            with path.open("w", encoding="utf-8") as log:
                process = subprocess.Popen(command, cwd=ROOT, env=environment,
                                           stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                                           text=True, encoding="utf-8", errors="replace")
                for line in process.stdout:
                    log.write(line)
                    log.flush()
                    print(line, end="", flush=True)
                if process.wait():
                    raise RuntimeError(f"Benchmark failed: {path}")
            matches = ROW.findall(path.read_text(encoding="utf-8"))
            if len(matches) != 2 or {r[0] for r in matches} != {"GenerateNormalArea", "LoadSavedNormalArea"}:
                raise RuntimeError(f"Expected both pipeline measurements: {path}")
            for name, ns, chunks, allocated, allocations in matches:
                if float(chunks) != 81:
                    raise RuntimeError(f"Workload changed: {path}")
                trials.append(dict(pair=pair, variant=label, benchmark=name,
                                   ns_per_op=int(ns), bytes_per_op=int(allocated),
                                   allocs_per_op=int(allocations)))
    require_clean()
    summary = {}
    for name in ("GenerateNormalArea", "LoadSavedNormalArea"):
        group = {}
        for label in ("before", "after"):
            rows = [r for r in trials if r["benchmark"] == name and r["variant"] == label]
            group[label] = {key: statistics.median(r[key] for r in rows)
                            for key in ("ns_per_op", "bytes_per_op", "allocs_per_op")}
        group["time_reduction_percent"] = (1 - group["after"]["ns_per_op"] / group["before"]["ns_per_op"]) * 100
        group["allocation_reduction_percent"] = (1 - group["after"]["bytes_per_op"] / group["before"]["bytes_per_op"]) * 100
        summary[name] = group
    (output / "results.json").write_text(json.dumps(dict(trials=trials, summary=summary), indent=2) + "\n", encoding="utf-8")
    print(json.dumps(summary, indent=2), flush=True)


if __name__ == "__main__":
    main()
