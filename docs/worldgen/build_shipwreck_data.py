"""Extract all 20 original Minecraft Java 1.21.1 shipwreck templates unchanged."""
import argparse
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/shipwreck_data"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
        version = json.loads(bundle.read("version.json"))
        if version["id"] != "1.21.1" or version["world_version"] != 3955:
            raise ValueError("expected Minecraft Java 1.21.1 server JAR")
        inner = bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar")
    prefix = "data/minecraft/structure/shipwreck/"
    with zipfile.ZipFile(io.BytesIO(inner)) as jar:
        expected = {name.removeprefix(prefix): jar.read(name) for name in jar.namelist() if name.startswith(prefix) and name.endswith(".nbt")}
    if len(expected) != 20:
        raise ValueError("expected exactly 20 original shipwreck templates")
    if args.check:
        if {path.name for path in OUTPUT.iterdir()} != set(expected):
            raise SystemExit("unexpected shipwreck asset inventory")
        for name, data in expected.items():
            if (OUTPUT / name).read_bytes() != data:
                raise SystemExit("original shipwreck asset differs: " + name)
        print("verified 20 original shipwreck templates")
    else:
        OUTPUT.mkdir(parents=True, exist_ok=True)
        for name, data in expected.items():
            (OUTPUT / name).write_bytes(data)
        print("extracted 20 original shipwreck templates")


if __name__ == "__main__":
    main()
