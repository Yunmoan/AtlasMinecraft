"""Independent 1.21.1 terrain checks for selected mountain/plateau/cave chunks.

Java is used only to regenerate reference vectors, never by Atlas at runtime.
Pass --cases with seed/cx/cz rows to reproduce a reported location.
"""
import argparse
import json
import os
import shutil
import subprocess
import tempfile
import zipfile
from pathlib import Path

from _bootstrap import common

HERE = Path(__file__).resolve().parent
FIXTURES = HERE.parents[1] / "testdata/worldgen"
PROBE_DIR = HERE / "probes"
ROOT = common.ROOT

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--update", action="store_true")
    parser.add_argument("--cases", type=Path, default=FIXTURES / "terrain_cases_1.21.1.txt")
    parser.add_argument("--output-dir", type=Path, default=FIXTURES)
    args = parser.parse_args()
    args.output_dir.mkdir(parents=True, exist_ok=True)
    common.java_tools()
    with tempfile.TemporaryDirectory(prefix="atlas-terrain-ref-") as temp:
        directory = Path(temp)
        jars = common.reference_jars(directory)
        inner = jars[0]
        classpath = os.pathsep.join(map(str, jars))
        subprocess.run([shutil.which("javac"), "-proc:none", "-cp", classpath, "-d", temp,
                        str(PROBE_DIR / "NoiseChunkProbe.java"), str(PROBE_DIR / "SurfaceChunkProbe.java")], check=True)
        classpath = os.pathsep.join([temp, classpath])
        noise = directory / "noise.txt"
        tasks = [("NoiseChunkProbe", [str(args.cases.resolve())], "terrain_noise_1.21.1.txt"),
                 ("SurfaceChunkProbe", [str(noise), "natural"], "terrain_surface_1.21.1.txt"),
                 ("SurfaceChunkProbe", [str(noise), "carved", str(inner)], "terrain_carved_1.21.1.txt")]
        for probe, parameters, name in tasks:
            result = subprocess.run([shutil.which("java"), "-Xmx2G", "-cp", classpath, f"atlasprobe.{probe}", *parameters],
                                    capture_output=True, cwd=temp)
            if result.returncode:
                raise SystemExit(f"{probe} failed:\n{result.stderr.decode('utf-8', errors='replace')}\n{result.stdout.decode('utf-8', errors='replace')[-2000:]}")
            actual = result.stdout.decode("utf-8").replace("\r\n", "\n")
            if probe == "NoiseChunkProbe":
                noise.write_text(actual, encoding="utf-8")
            fixture = args.output_dir / name
            if args.update:
                fixture.write_text(actual, encoding="utf-8", newline="\n")
                print(f"updated {name}", flush=True)
            elif actual != fixture.read_text(encoding="utf-8"):
                raise SystemExit(f"reference mismatch: {name}")
            else:
                print(f"verified {name}", flush=True)

if __name__ == "__main__":
    main()
