"""Extract the JAR's substrate tags for the supported vegetation features."""
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
with zipfile.ZipFile(ROOT / "jar/server.jar") as bundle:
    with zipfile.ZipFile(io.BytesIO(bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar"))) as jar:
        def resolve(name):
            result = set()
            for value in json.loads(jar.read(f"data/minecraft/tags/block/{name}.json"))["values"]:
                if value.startswith("#minecraft:"):
                    result.update(resolve(value.removeprefix("#minecraft:")))
                else:
                    result.add(value)
            return sorted(result)
        data = {name: resolve(name) for name in ("dead_bush_may_place_on", "dirt", "snow_layer_cannot_survive_on", "snow_layer_can_survive_on", "dripstone_replaceable_blocks", "moss_replaceable", "lush_ground_replaceable", "small_dripleaf_placeable", "azalea_grows_on", "azalea_root_replaceable", "features_cannot_replace", "lava_pool_stone_cannot_replace", "geode_invalid_blocks", "bamboo_plantable_on", "mushroom_grow_block")}
(ROOT / "internal/worldgen/vegetation_tags.json").write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
