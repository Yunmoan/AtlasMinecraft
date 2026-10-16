from pathlib import Path
import os
import json
import re
import subprocess

root = Path(__file__).resolve().parents[2]
target = root / "artifacts/worldgen-probes"
classpath = os.pathsep.join([str(target), str(root / "vanilla/server-inner.jar")] +
                          [str(p) for p in (root / "vanilla/libraries").rglob("*.jar")])
subprocess.run(["javac", "-proc:none", "-cp", classpath, "-d", str(target),
                str(root / "docs/worldgen/InitialAnimalSpawnProbe.java")], check=True, cwd=root)
fixture = root / "docs/worldgen/initial_animals_1.21.1.txt"
temporary = fixture.with_suffix(".txt.tmp")
with temporary.open("w", encoding="utf-8") as output:
    subprocess.run(["java", "-cp", classpath, "atlasprobe.InitialAnimalSpawnProbe"],
                   check=True, stdout=output, cwd=root)
temporary.replace(fixture)

biomes, types, states, defaults = {}, {}, [], {}
for line in fixture.read_text().splitlines():
    parts = line.split()
    if parts[0] == "creature_settings":
        biomes[parts[1]] = {"probability": float(parts[2]), "entries": [
            {"id": name, "weight": int(weight), "min": int(low), "max": int(high)}
            for name, low, high, weight in re.findall(r"(minecraft:\w+)\*\((\d+)-(\d+)\):(\d+)", line)]}
    elif parts[0] == "animal_type":
        types[parts[1]] = {"entity_id": int(parts[2]), "width": float(parts[3]), "height": float(parts[4]), "summon": parts[5] == "true"}
    elif parts[0] == "animal_state":
        states.append([int(parts[2]), int(parts[3])])
    elif parts[0] == "animal_default":
        defaults[parts[1]] = parts[3]
assert len(biomes) == 64 and len(types) == 6 and len(states) == 26684
data = root / "internal/worldgen/initial_animal_data.json"
temp_data = data.with_suffix(".json.tmp")
temp_data.write_text(json.dumps({"biomes": biomes, "types": types, "states": states}, separators=(",", ":")))
temp_data.replace(data)
animal_defaults = root / "internal/world/animal_defaults_1.21.1.txt"
temporary = animal_defaults.with_suffix(".txt.tmp")
temporary.write_text("".join(f"{name} {data}\n" for name, data in defaults.items()), encoding="utf-8")
temporary.replace(animal_defaults)
