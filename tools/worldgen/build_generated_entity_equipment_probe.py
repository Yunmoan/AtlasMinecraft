from pathlib import Path
import os,subprocess
from _bootstrap import common

root = common.ROOT;target=root/'artifacts/worldgen-probes'
cp=common.reference_classpath(str(target))
subprocess.run(['javac','-proc:none','-cp',cp,'-d',str(target),str(root/'tools/worldgen/probes/InitialAnimalSpawnProbe.java'),str(root/'tools/worldgen/probes/GeneratedEntityPacketProbe.java'),str(root/'tools/worldgen/probes/GeneratedEntityEquipmentProbe.java')],check=True,cwd=root)
fixture=root/'testdata/worldgen/generated_entity_equipment_1.21.1.txt';tmp=fixture.with_suffix('.txt.tmp')
with tmp.open('w',encoding='utf-8') as output:subprocess.run(['java','-cp',cp,'atlasprobe.GeneratedEntityEquipmentProbe'],check=True,cwd=root,stdout=output)
tmp.replace(fixture)
