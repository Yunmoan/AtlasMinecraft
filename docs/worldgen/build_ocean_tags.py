"""Extract ordered coral tags from the official 1.21.1 JAR."""
import argparse
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/ocean_tags.json"

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true")
    args = parser.parse_args()
    with zipfile.ZipFile(ROOT / "vanilla/server-inner.jar") as jar:
        def expand(name):
            result = []
            for entry in json.loads(jar.read(f"data/minecraft/tags/block/{name}.json"))["values"]:
                value = entry if isinstance(entry, str) else entry["id"]
                for block in expand(value[11:]) if value.startswith("#minecraft:") else [value]:
                    if block not in result:
                        result.append(block)
            return result
        tags = {name: expand(name) for name in ("coral_blocks", "corals", "wall_corals")}
    text = json.dumps(tags, indent=2, sort_keys=True) + "\n"
    if args.check:
        if OUTPUT.read_text(encoding="utf-8") != text:
            raise SystemExit("ordered coral tags differ from the official JAR")
        print("verified ocean_tags.json")
    else:
        OUTPUT.write_text(text, encoding="utf-8", newline="\n")
        print("extracted ocean_tags.json")

if __name__ == "__main__":
    main()
