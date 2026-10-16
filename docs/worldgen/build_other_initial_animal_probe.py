from pathlib import Path
import os
import subprocess
import json
import zipfile

root=Path(__file__).resolve().parents[2]
target=root / "artifacts/worldgen-probes"
cp=os.pathsep.join([str(target),str(root / "vanilla/server-inner.jar")]+[str(p) for p in (root / "vanilla/libraries").rglob("*.jar")])
subprocess.run(["javac","-proc:none","-cp",cp,"-d",str(target),str(root / "docs/worldgen/InitialAnimalSpawnProbe.java"),str(root / "docs/worldgen/OtherInitialAnimalProbe.java")],check=True,cwd=root)
fixture=root / "docs/worldgen/other_initial_animals_1.21.1.txt"
temporary=fixture.with_suffix(".txt.tmp")
with temporary.open("w",encoding="utf-8") as output:
    subprocess.run(["java","-cp",cp,"atlasprobe.OtherInitialAnimalProbe"],check=True,stdout=output,cwd=root)
temporary.replace(fixture)

types, states, defaults, biome_tags, registries, placement_lower = {}, [], {}, {}, {}, []
for line in fixture.read_text().splitlines():
    p=line.split()
    if p[0]=='animal_type':
        types[p[1]]={'entity_id':int(p[2]),'width':float(p[3]),'height':float(p[4]),'summon':p[5]=='true','heightmap':p[6]}
    elif p[0]=='animal_state':
        states.append([int(p[2]),int(p[3])])
    elif p[0]=='animal_path':
        placement_lower.append(p[2]=='true')
    elif p[0]=='animal_default':
        defaults[p[1]]=p[3]
    elif p[0]=='animal_biome_tag':
        biome_tags[p[1]]=p[2].split(',') if len(p)>2 else []
    elif p[0]=='animal_registry':
        registries.setdefault(p[1],{})[p[2]]=int(p[3])
assert len(types)==12 and len(states)==26684
with zipfile.ZipFile(root/'vanilla/server-inner.jar') as jar:
    def members(name):
        result=[]
        for value in json.loads(jar.read('data/minecraft/tags/block/'+name+'.json'))['values']:
            if value.startswith('#minecraft:'):result+=members(value.removeprefix('#minecraft:'))
            else:result.append(value)
        return result
    block_tags={name:members(name) for name in ['animals_spawnable_on','armadillo_spawnable_on','foxes_spawnable_on','frogs_spawnable_on','goats_spawnable_on','mooshrooms_spawnable_on','parrots_spawnable_on','polar_bears_spawnable_on_alternate','rabbits_spawnable_on','wolves_spawnable_on']}
data=root/'internal/worldgen/other_initial_animal_data.json'
temp=data.with_suffix('.json.tmp');temp.write_text(json.dumps({'types':types,'states':states,'placement_lower':placement_lower,'biome_tags':biome_tags,'block_tags':block_tags,'registries':registries},separators=(',',':')));temp.replace(data)
data=root/'internal/world/other_animal_defaults_1.21.1.txt'
temp=data.with_suffix('.txt.tmp');temp.write_text(''.join(f'{name} {nbt}\n' for name,nbt in defaults.items()),encoding='utf-8');temp.replace(data)
