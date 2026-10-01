"""Extract Minecraft 1.21.1 built-in feature definitions for the Go port."""

import argparse
import io
import json
import zipfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/feature_data.json"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()

    with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
        version = json.loads(bundle.read("version.json"))
        if version["id"] != "1.21.1" or version["world_version"] != 3955:
            raise ValueError("expected Minecraft Java 1.21.1 server JAR")
        server = bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar")

    definitions = {}
    with zipfile.ZipFile(io.BytesIO(server)) as jar:
        for kind in ("placed_feature", "configured_feature"):
            prefix = f"data/minecraft/worldgen/{kind}/"
            definitions[kind] = {
                "minecraft:" + name[len(prefix) : -5]: json.loads(jar.read(name))
                for name in jar.namelist()
                if name.startswith(prefix) and name.endswith(".json")
            }
        ore_tags = ("base_stone_overworld", "stone_ore_replaceables",
                    "deepslate_ore_replaceables", "base_stone_nether")
        definitions["ore_replaceable_tags"] = {
            "minecraft:" + name: json.loads(jar.read(
                f"data/minecraft/tags/block/{name}.json"))["values"]
            for name in ore_tags
        }
    result = json.dumps(definitions, indent=2, sort_keys=True, ensure_ascii=False) + "\n"
    if args.check:
        if OUTPUT.read_text(encoding="utf-8") != result:
            raise SystemExit(f"feature definitions differ from {OUTPUT}")
        print(f"verified {OUTPUT}")
    else:
        OUTPUT.write_text(result, encoding="utf-8", newline="\n")
        print(f"extracted {len(definitions['placed_feature'])} placed and "
              f"{len(definitions['configured_feature'])} configured features to {OUTPUT}")


if __name__ == "__main__":
    main()
