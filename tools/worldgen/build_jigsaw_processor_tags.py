"""Extract/check original block tags referenced by all 1.21.1 Jigsaw processors."""
import argparse
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/jigsaw_processor_tags_1.21.1.json"

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    registry = json.loads((ROOT / "internal/worldgen/jigsaw_data/registry.json").read_text())
    tags = set()
    def visit(value):
        if isinstance(value, dict):
            for key, item in value.items():
                if key == "tag" and isinstance(item, str):
                    tags.add(item.removeprefix("minecraft:"))
                if key in ("rottable_blocks", "value") and isinstance(item, str) and item.startswith("#"):
                    tags.add(item.removeprefix("#minecraft:"))
                visit(item)
        elif isinstance(value, list):
            for item in value:
                visit(item)
    visit(registry["processors"])
    with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
        version = json.loads(bundle.read("version.json"))
        if version["id"] != "1.21.1" or version["world_version"] != 3955:
            raise ValueError("expected original Minecraft Java 1.21.1")
        with zipfile.ZipFile(io.BytesIO(bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar"))) as jar:
            def resolve(name):
                result = set()
                for item in json.loads(jar.read(f"data/minecraft/tags/block/{name}.json"))["values"]:
                    value = item if isinstance(item, str) else item["id"]
                    if value.startswith("#"):
                        result.update(resolve(value.removeprefix("#minecraft:")))
                    else:
                        result.add(value)
                return sorted(result)
            result = {"minecraft:" + name: resolve(name) for name in sorted(tags)}
    data = (json.dumps({"version": 3955, "tags": result}, sort_keys=True, separators=(",", ":")) + "\n").encode()
    if args.check:
        if OUTPUT.read_bytes() != data:
            raise SystemExit("original Jigsaw processor block tags differ")
        print(f"verified {len(tags)} original Jigsaw processor block tags")
    else:
        OUTPUT.write_bytes(data)
        print(f"extracted {len(tags)} original Jigsaw processor block tags")

if __name__ == "__main__":
    main()
