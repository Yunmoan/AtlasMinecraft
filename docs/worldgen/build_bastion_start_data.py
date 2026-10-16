"""Extract/check the original four Bastion start-pool assets and pool definition."""
import argparse
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUTPUT = ROOT / "internal/worldgen/bastion_start_data"


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
        pool = jar.read("data/minecraft/worldgen/template_pool/bastion/starts.json")
        expected = {"starts.json": pool}
        for item in json.loads(pool)["elements"]:
            name = item["element"]["location"].removeprefix("minecraft:")
            expected[name + ".nbt"] = jar.read("data/minecraft/structure/" + name + ".nbt")
    if args.check:
        if {p.relative_to(OUTPUT).as_posix() for p in OUTPUT.rglob("*") if p.is_file()} != set(expected):
            raise SystemExit("unexpected Bastion start asset inventory")
        for name, data in expected.items():
            if (OUTPUT / name).read_bytes() != data:
                raise SystemExit("original Bastion start differs: " + name)
        print("verified four original Bastion start templates and pool")
    else:
        for name, data in expected.items():
            destination = OUTPUT / name
            destination.parent.mkdir(parents=True, exist_ok=True)
            destination.write_bytes(data)
        print("extracted four original Bastion start templates and pool")


if __name__ == "__main__":
    main()
