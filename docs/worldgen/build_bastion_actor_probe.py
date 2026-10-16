"""Run actual authored Bastion actor loading/finalization with the official JAR."""
from pathlib import Path
import os
import subprocess
import sys

ROOT=Path(__file__).resolve().parents[2]
HERE=Path(__file__).resolve().parent
TARGET=ROOT/'artifacts/worldgen-probes'
sys.path.insert(0,str(HERE))
from verify_structure_reference import atomic_write

classpath=os.pathsep.join([str(TARGET),str(ROOT/'vanilla/server-inner.jar')]+[str(p) for p in (ROOT/'vanilla/libraries').rglob('*.jar')])
subprocess.run(['javac','-proc:none','-cp',classpath,'-d',str(TARGET),*[str(HERE/name) for name in ('InitialAnimalSpawnProbe.java','GeneratedEntityProbe.java','BastionActorProbe.java')]],check=True,cwd=ROOT)
process=subprocess.run(['java','-cp',classpath,'atlasprobe.BastionActorProbe'],check=True,capture_output=True,cwd=ROOT)
fixture=HERE/'structure_bastion_actors_1.21.1.txt'
actual=process.stdout.decode().replace('\r\n','\n')
atomic_write(fixture,actual.encode())
defaults=[]
for line in actual.splitlines():
    fields=line.split()
    if fields[0]=='bastion_actor_default':defaults.append(f'{fields[1]} {fields[3]}\n')
assert len(defaults)==3
atomic_write(ROOT/'internal/worldgen/bastion_actor_defaults_1.21.1.txt',''.join(defaults).encode())
atomic_write(ROOT/'internal/world/bastion_actor_defaults_1.21.1.txt',''.join(defaults).encode())
print(f'published {len(actual.splitlines())} official Bastion actor records')
