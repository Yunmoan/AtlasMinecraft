from pathlib import Path
import os,subprocess
root=Path(__file__).resolve().parents[2];target=root/'artifacts/worldgen-probes'
cp=os.pathsep.join([str(target),str(root/'vanilla/server-inner.jar')]+[str(p) for p in (root/'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac','-proc:none','-cp',cp,'-d',str(target),str(root/'docs/worldgen/InitialAnimalSpawnProbe.java'),str(root/'docs/worldgen/GeneratedEntityProbe.java'),str(root/'docs/worldgen/AtlasProbeGameEventDispatcher.java'),str(root/'docs/worldgen/DrownedActorProbe.java')],check=True,cwd=root)
fixture=root/'docs/worldgen/structure_drowned_actors_1.21.1.txt';tmp=fixture.with_suffix('.txt.tmp')
with tmp.open('w',encoding='utf-8') as output:subprocess.run(['java','-cp',cp,'atlasprobe.DrownedActorProbe'],check=True,cwd=root,stdout=output)
tmp.replace(fixture)
defaults=[line.split() for line in fixture.read_text().splitlines() if line.startswith('drowned_actor_default ')]
assert len(defaults)==1
for folder in ['worldgen','world']:
 path=root/'internal'/folder/'drowned_actor_defaults_1.21.1.txt';tmp=path.with_suffix('.txt.tmp');tmp.write_text(f'{defaults[0][1]} {defaults[0][3]}\n');tmp.replace(path)
