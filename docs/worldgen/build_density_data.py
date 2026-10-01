"""Extract the 1.21.1 density graph and noise settings from jar/server.jar.

The generated JSON is embedded by Go; the server JAR is never read at runtime.
Run with --check to compare the checked-in data without changing it.
"""
import argparse
import io
import json
import zipfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal" / "worldgen" / "density_data.json"


def extract():
    with zipfile.ZipFile(ROOT / "jar" / "server.jar") as bundle:
        version = json.loads(bundle.read("version.json"))
        if version["id"] != "1.21.1" or version["world_version"] != 3955:
            raise ValueError("expected Minecraft Java 1.21.1 / DataVersion 3955")
        server_bytes = bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar")

    groups = {}
    with zipfile.ZipFile(io.BytesIO(server_bytes)) as server:
        for group in ("density_function", "noise_settings"):
            prefix = f"data/minecraft/worldgen/{group}/"
            groups[group] = {
                "minecraft:" + path[len(prefix):-5]: json.loads(server.read(path))
                for path in server.namelist()
                if path.startswith(prefix) and path.endswith(".json")
            }
    if len(groups["density_function"]) != 35 or len(groups["noise_settings"]) != 7:
        raise ValueError("unexpected 1.21.1 built-in worldgen registry size")
    return (json.dumps(groups, sort_keys=True, separators=(",", ":")) + "\n").encode()


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    result = extract()
    if args.check:
        if OUTPUT.read_bytes() != result:
            raise SystemExit(f"extracted data differs from {OUTPUT}")
        print("verified density_data.json against server.jar")
    else:
        OUTPUT.write_bytes(result)
        print(f"wrote {len(result)} bytes to {OUTPUT}")
