from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
target = root / "artifacts/worldgen-probes"
classpath = os.pathsep.join([str(target), str(root / "vanilla/server-inner.jar")] +
                          [str(p) for p in (root / "vanilla/libraries").rglob("*.jar")])
subprocess.run(["javac", "-proc:none", "-cp", classpath, "-d", str(target),
                str(root / "docs/worldgen/AtlasEntityStructureManager.java"),
                str(root / "docs/worldgen/GeneratedEntityProbe.java")], check=True, cwd=root)
fixture = root / "docs/worldgen/generated_entities_1.21.1.txt"
temporary = fixture.with_suffix(".txt.tmp")
with temporary.open("w", encoding="utf-8") as output:
    subprocess.run(["java", "-cp", classpath, "atlasprobe.GeneratedEntityProbe"],
                   check=True, stdout=output, cwd=root)
temporary.replace(fixture)
defaults = {}
for line in (root / "docs/worldgen/generated_entities_1.21.1.txt").read_text().splitlines():
    parts = line.split()
    if parts and parts[0] == "entity_default" and parts[1] not in defaults:
        defaults[parts[1]] = parts[3]
(root / "internal/world/entity_defaults_1.21.1.txt").write_text(
    "".join(f"{name} {data}\n" for name, data in defaults.items()), encoding="utf-8")
