"""Extract/check unmodified 1.21.1 Jigsaw pools, processors and their templates."""
import argparse
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/jigsaw_data"


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
        pools = {}
        processors = {}
        templates = set()
        def collect(element):
            if element["element_type"] in ("minecraft:single_pool_element", "minecraft:legacy_single_pool_element"):
                templates.add(element["location"].removeprefix("minecraft:"))
            elif element["element_type"] == "minecraft:list_pool_element":
                for item in element["elements"]:
                    collect(item)
        for name in jar.namelist():
            prefix = "data/minecraft/worldgen/template_pool/"
            if name.startswith(prefix) and name.endswith(".json"):
                pools["minecraft:" + name[len(prefix):-5]] = json.loads(jar.read(name))
                for item in pools["minecraft:" + name[len(prefix):-5]]["elements"]:
                    collect(item["element"])
            prefix = "data/minecraft/worldgen/processor_list/"
            if name.startswith(prefix) and name.endswith(".json"):
                processors["minecraft:" + name[len(prefix):-5]] = json.loads(jar.read(name))
        expected = {}
        missing = []
        for name in sorted(templates):
            resource = "data/minecraft/structure/" + name + ".nbt"
            if resource in jar.namelist():
                expected["templates/" + name + ".nbt"] = jar.read(resource)
            else:
                # Vanilla's getOrCreate returns an empty template for these
                # built-in pool entries; keep that absence explicit.
                missing.append("minecraft:" + name)
        expected["registry.json"] = (json.dumps({"version": 3955, "pools": pools, "processors": processors, "missing_templates": missing},sort_keys=True,separators=(",", ":")) + "\n").encode()
    if args.check:
        if {p.relative_to(OUTPUT).as_posix() for p in OUTPUT.rglob("*") if p.is_file()} != set(expected):
            raise SystemExit("unexpected Jigsaw asset inventory")
        for name, data in expected.items():
            if (OUTPUT / name).read_bytes() != data:
                raise SystemExit("original Jigsaw asset differs: " + name)
        print(f"verified {len(pools)} original Jigsaw pools, {len(processors)} processors, {len(templates)} templates")
    else:
        for name, data in expected.items():
            destination = OUTPUT / name
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_bytes(data)
        print(f"extracted {len(pools)} original Jigsaw pools, {len(processors)} processors, {len(templates)} templates")


if __name__ == "__main__":
    main()
