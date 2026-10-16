"""Extract the 13 original 1.21.1 ruined-portal NBT templates unchanged."""
import argparse
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/ruined_portal_data"


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
        names = [f"portal_{index}" for index in range(1, 11)]
        names += [f"giant_portal_{index}" for index in range(1, 4)]
        expected = {name + ".nbt": jar.read("data/minecraft/structure/ruined_portal/" + name + ".nbt") for name in names}
    if args.check:
        if {p.name for p in OUTPUT.iterdir()} != set(expected):
            raise SystemExit("unexpected ruined portal asset inventory")
        for name, data in expected.items():
            if (OUTPUT / name).read_bytes() != data:
                raise SystemExit("original ruined portal asset differs: " + name)
        print("verified 13 original ruined portal templates")
    else:
        OUTPUT.mkdir(parents=True, exist_ok=True)
        for name, data in expected.items():
            (OUTPUT / name).write_bytes(data)
        print("extracted 13 original ruined portal templates")


if __name__ == "__main__":
    main()
