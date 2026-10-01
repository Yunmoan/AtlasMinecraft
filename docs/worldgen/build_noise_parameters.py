"""Extract vanilla 1.21.1 noise parameters from the supplied server JAR."""
import io
import json
import zipfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
with zipfile.ZipFile(ROOT / "jar" / "server.jar") as bundle:
    version = json.loads(bundle.read("version.json"))
    if version["id"] != "1.21.1" or version["world_version"] != 3955:
        raise ValueError("expected the Java 1.21.1 server JAR")
    server_bytes = bundle.read("META-INF/versions/1.21.1/server-1.21.1.jar")

parameters = {}
prefix = "data/minecraft/worldgen/noise/"
with zipfile.ZipFile(io.BytesIO(server_bytes)) as server:
    for path in server.namelist():
        if path.startswith(prefix) and path.endswith(".json"):
            name = "minecraft:" + path[len(prefix):-5]
            parameters[name] = json.loads(server.read(path))

output = ROOT / "internal" / "worldgen" / "noise_parameters.json"
output.write_text(json.dumps(parameters, sort_keys=True, separators=(",", ":")) + "\n", encoding="utf-8")
print(f"wrote {len(parameters)} vanilla noise parameters to {output}")
