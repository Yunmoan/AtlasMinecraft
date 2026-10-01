"""Extract the block tags used by the 1.21.1 tree port (JAR is test-only)."""
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
    import io
    with zipfile.ZipFile(io.BytesIO(bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar"))) as jar:
        def resolve(name):
            result = set()
            for value in json.loads(jar.read(f"data/minecraft/tags/block/{name}.json"))["values"]:
                if value.startswith("#minecraft:"):
                    result.update(resolve(value.removeprefix("#minecraft:")))
                else:
                    result.add(value)
            return sorted(result)
        data = {name: resolve(name) for name in ("replaceable_by_trees", "logs", "leaves")}
(ROOT / "internal/worldgen/tree_tags.json").write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
