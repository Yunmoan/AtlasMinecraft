"""Run the existing native terrain probes for a small explicit chunk set.

This development helper writes diagnostic artifacts only. It does not update
checked-in reference fixtures or use Java in the server runtime.
"""

import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import zipfile


ROOT = Path(__file__).resolve().parents[2]
HERE = Path(__file__).resolve().parent


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--seed", type=int, required=True)
    parser.add_argument("--cx", type=int)
    parser.add_argument("--cz", type=int)
    parser.add_argument("--chunk", type=int, nargs=2, action="append", metavar=("CX", "CZ"),
                        help="explicit chunk; repeat for at most 9 chunks")
    parser.add_argument("--radius", type=int, choices=(0, 1), default=0)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    if args.chunk:
        if args.cx is not None or args.cz is not None or args.radius != 0:
            parser.error("--chunk cannot be combined with --cx, --cz, or --radius")
        chunks = [tuple(chunk) for chunk in args.chunk]
        if len(chunks) > 9 or len(set(chunks)) != len(chunks):
            parser.error("--chunk requires at most 9 distinct chunks")
    else:
        if args.cx is None or args.cz is None:
            parser.error("supply --cx and --cz, or one or more --chunk CX CZ")
        chunks = [
            (cx, cz)
            for cx in range(args.cx - args.radius, args.cx + args.radius + 1)
            for cz in range(args.cz - args.radius, args.cz + args.radius + 1)
        ]
    java, javac = shutil.which("java"), shutil.which("javac")
    if not java or not javac:
        raise SystemExit("JDK 21 java and javac must be on PATH")
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=True)
    cases = output / "cases.txt"
    cases.write_text("".join(
        f"{args.seed} {cx} {cz}\n"
        for cx, cz in chunks
    ), encoding="utf-8")
    with tempfile.TemporaryDirectory(prefix="atlas-terrain-diag-") as temp:
        temporary = Path(temp)
        with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
            version = json.loads(bundle.read("version.json"))
            if version["id"] != "1.21.1" or version["world_version"] != 3955:
                raise ValueError("expected Minecraft Java 1.21.1 server JAR")
            inner = temporary / "server-1.21.1.jar"
            inner.write_bytes(bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar"))
            libraries = []
            for member in bundle.namelist():
                if member.startswith("META-INF/libraries/") and member.endswith(".jar"):
                    path = temporary / f"library-{len(libraries)}.jar"
                    path.write_bytes(bundle.read(member))
                    libraries.append(path)
        classpath = os.pathsep.join(map(str, [inner, *libraries]))
        names = ("NoiseChunkProbe", "SurfaceChunkProbe", "CarverChunkProbe")
        subprocess.run([javac, "-proc:none", "-cp", classpath, "-d", temp,
                        *(str(HERE / f"{name}.java") for name in names)], check=True)
        classpath = os.pathsep.join([temp, classpath])
        commands = (
            ("NoiseChunkProbe", [str(cases)], "noise.txt", "chunk "),
            ("SurfaceChunkProbe", [str(output / "noise.txt"), "natural"], "surface.txt", "surface_chunk "),
            ("CarverChunkProbe", [str(output / "noise.txt"), str(inner)], "carved.txt", "carved_chunk "),
        )
        for name, probe_args, filename, prefix in commands:
            result = subprocess.run([java, "-cp", classpath, f"atlasprobe.{name}", *probe_args],
                                    capture_output=True, cwd=temporary, check=False)
            (output / f"{name}.stderr.txt").write_bytes(result.stderr)
            if result.returncode:
                raise SystemExit(f"{name} failed: {result.stderr.decode('utf-8', errors='replace')}"
                                 f"\n{result.stdout.decode('utf-8', errors='replace')[-4000:]}")
            lines = result.stdout.decode("utf-8").splitlines()
            selected = [line for line in lines if line.startswith(prefix)
                        or name == "NoiseChunkProbe" and line.startswith("density ")]
            (output / filename).write_text("\n".join(selected) + "\n", encoding="utf-8")
            print(f"wrote {filename}: {sum(line.startswith(prefix) for line in selected)} chunks", flush=True)
    print(f"Set ATLAS_TERRAIN_DIAG_DIR={output} and run "
          "go test ./internal/worldgen -run '^TestTerrainStageDiagnostic$' -count=1 -v", flush=True)


if __name__ == "__main__":
    main()
