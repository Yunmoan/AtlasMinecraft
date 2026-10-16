"""Reproduce original 1.21.1 stronghold blocks in an isolated Java cwd."""
from pathlib import Path
import argparse
import os
import subprocess
import tempfile
import sys

ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent
TARGET = ROOT / "artifacts/stronghold-v12"
TARGET.mkdir(parents=True, exist_ok=True)
sys.path.insert(0, str(HERE))
from verify_structure_reference import atomic_write

parser = argparse.ArgumentParser()
parser.add_argument("--check", action="store_true")
args = parser.parse_args()
classpath = os.pathsep.join([str(TARGET), str(ROOT / "vanilla/server-inner.jar")] + [str(p) for p in (ROOT / "vanilla/libraries").rglob("*.jar")])
with tempfile.TemporaryDirectory(prefix="atlas-stronghold-blocks-") as temporary:
    sources = [str(HERE / name) for name in ("IglooProbeSupport.java", "RuinedPortalProbeSupport.java", "RuinedPortalStartsProbe.java", "StrongholdBlocksProbe.java")]
    compiler = subprocess.run(["javac", "-proc:none", "-cp", classpath, "-d", str(TARGET), *sources], cwd=temporary, capture_output=True)
    atomic_write(TARGET / "blocks-javac.txt", compiler.stdout + compiler.stderr)
    compiler.check_returncode()
    process = subprocess.run(["java", "-cp", classpath, "atlasprobe.StrongholdBlocksProbe"], cwd=temporary, capture_output=True)
    atomic_write(TARGET / "blocks.stderr.txt", process.stderr)
    process.check_returncode()
    atomic_write(TARGET / "blocks.stdout.txt", process.stdout)
rows = [line for line in process.stdout.splitlines() if line.startswith(b"stronghold_block ")]
assert len(rows) == 8064, len(rows)
data = b"\n".join(rows) + b"\n"
fixture = HERE / "structure_stronghold_blocks_1.21.1.txt"
if args.check:
    assert fixture.read_bytes() == data, "stronghold blocks oracle changed"
else:
    atomic_write(fixture, data)
print(f"{'checked' if args.check else 'published'} {len(rows)} original stronghold layout/start records")
