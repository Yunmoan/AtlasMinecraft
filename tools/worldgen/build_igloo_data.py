"""Extract original 1.21.1 igloo structure NBT unchanged; runtime is pure Go."""
import argparse
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/igloo_data"

def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
        version = json.loads(bundle.read("version.json"))
        if version["id"] != "1.21.1" or version["world_version"] != 3955:
            raise ValueError("expected Minecraft Java 1.21.1 server JAR")
        inner = bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar")
    with zipfile.ZipFile(io.BytesIO(inner)) as jar:
        expected = {name + ".nbt": jar.read("data/minecraft/structure/igloo/" + name + ".nbt") for name in ("top", "middle", "bottom")}
    if args.check:
        if {p.name for p in OUTPUT.iterdir()} != set(expected): raise SystemExit("unexpected igloo asset inventory")
        for name, data in expected.items():
            if (OUTPUT / name).read_bytes() != data: raise SystemExit("original igloo asset differs: " + name)
        print("verified 3 original igloo templates")
    else:
        OUTPUT.mkdir(parents=True, exist_ok=True)
        for name, data in expected.items(): (OUTPUT / name).write_bytes(data)
        print("extracted 3 original igloo templates")

if __name__ == "__main__": main()
