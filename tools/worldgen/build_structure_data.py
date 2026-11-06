"""Extract original 1.21.1 structure placement and biome definitions.

These are prerequisites for structure starts, not generated structure blocks.
Atlas uses the extracted data without starting Java at runtime.
"""

import argparse
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/structure_data.json"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
        version = json.loads(bundle.read("version.json"))
        if version["id"] != "1.21.1" or version["world_version"] != 3955:
            raise ValueError("expected Minecraft Java 1.21.1 server JAR")
        inner = bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar")
    definitions = {}
    with zipfile.ZipFile(io.BytesIO(inner)) as jar:
        for kind in ("structure_set", "structure"):
            prefix = f"data/minecraft/worldgen/{kind}/"
            definitions[kind] = {
                "minecraft:" + name[len(prefix):-5]: json.loads(jar.read(name))
                for name in jar.namelist()
                if name.startswith(prefix) and name.endswith(".json")
            }
        prefix = "data/minecraft/tags/worldgen/biome/"
        definitions["biome_tags"] = {
            "minecraft:" + name[len(prefix):-5]: json.loads(jar.read(name))["values"]
            for name in jar.namelist()
            if name.startswith(prefix) and name.endswith(".json")
        }
    result = json.dumps(definitions, indent=2, sort_keys=True) + "\n"
    if args.check:
        if OUTPUT.read_text(encoding="utf-8") != result:
            raise SystemExit(f"structure definitions differ from {OUTPUT}")
        print(f"verified {OUTPUT}")
    else:
        OUTPUT.write_text(result, encoding="utf-8", newline="\n")
        print(f"extracted {len(definitions['structure_set'])} sets and "
              f"{len(definitions['structure'])} structures to {OUTPUT}")


if __name__ == "__main__":
    main()
