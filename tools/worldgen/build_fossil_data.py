"""Extract the original 1.21.1 fossil templates and processor lists.

The compressed NBT is embedded unchanged. No JAR or JVM is used at runtime.
"""
import argparse
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/fossil_data"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
        version = json.loads(bundle.read("version.json"))
        if version["id"] != "1.21.1" or version["world_version"] != 3955:
            raise ValueError("expected Minecraft Java 1.21.1 server JAR")
        server = bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar")
    expected = {}
    with zipfile.ZipFile(io.BytesIO(server)) as jar:
        for kind in ("spine", "skull"):
            for i in range(1, 5):
                for suffix in ("", "_coal"):
                    name = f"{kind}_{i}{suffix}.nbt"
                    expected[name] = jar.read("data/minecraft/structure/fossil/" + name)
        for name in ("fossil_rot", "fossil_coal", "fossil_diamonds"):
            expected[name + ".json"] = jar.read(
                f"data/minecraft/worldgen/processor_list/{name}.json")
    if args.check:
        if {p.name for p in OUTPUT.iterdir()} != set(expected):
            raise SystemExit("unexpected fossil asset inventory")
        for name, data in expected.items():
            if (OUTPUT / name).read_bytes() != data:
                raise SystemExit(f"fossil data differs: {name}")
        print(f"verified {len(expected)} original fossil assets")
    else:
        OUTPUT.mkdir(parents=True, exist_ok=True)
        for name, data in expected.items():
            (OUTPUT / name).write_bytes(data)
        print(f"extracted {len(expected)} original fossil assets")


if __name__ == "__main__":
    main()
