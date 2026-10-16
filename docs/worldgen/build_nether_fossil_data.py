"""Extract/check all 14 unchanged Minecraft 1.21.1 Nether fossil templates."""
import argparse
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/nether_fossil_data"

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
        version = json.loads(bundle.read("version.json"))
        if version["id"] != "1.21.1" or version["world_version"] != 3955:
            raise ValueError("expected Minecraft Java 1.21.1 server JAR")
        inner = bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar")
    with zipfile.ZipFile(io.BytesIO(inner)) as jar:
        expected = {f"fossil_{i}.nbt": jar.read(f"data/minecraft/structure/nether_fossils/fossil_{i}.nbt") for i in range(1, 15)}
    if args.check:
        if {p.name for p in OUTPUT.iterdir()} != set(expected):
            raise SystemExit("unexpected Nether fossil asset inventory")
        for name, data in expected.items():
            if (OUTPUT / name).read_bytes() != data:
                raise SystemExit("original Nether fossil differs: " + name)
        print("verified 14 original Nether fossil templates")
    else:
        OUTPUT.mkdir(parents=True, exist_ok=True)
        for name, data in expected.items():
            (OUTPUT / name).write_bytes(data)
        print("extracted 14 original Nether fossil templates")

if __name__ == "__main__":
    main()
