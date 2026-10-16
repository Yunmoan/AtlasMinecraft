"""Run actual original End City layouts/starts with full saved NBT and RNG."""
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
compiler=subprocess.run(['javac','-proc:none','-cp',classpath,'-d',str(TARGET),*[str(HERE/name) for name in ('IglooProbeSupport.java','RuinedPortalProbeSupport.java','RuinedPortalStartsProbe.java','EndCityStartsProbe.java')]],capture_output=True,cwd=ROOT)
atomic_write(TARGET/'end-city-starts-javac-v8.txt',compiler.stdout+compiler.stderr)
compiler.check_returncode()
process=subprocess.run(['java','-cp',classpath,'atlasprobe.EndCityStartsProbe'],capture_output=True,cwd=ROOT)
atomic_write(TARGET/'end-city-starts-v8.stderr.txt',process.stderr)
process.check_returncode()
atomic_write(TARGET/'end-city-starts-v8.stdout.txt',process.stdout)
rows=[line.decode('ascii') for line in process.stdout.splitlines() if line.startswith((b'end_city_layout ',b'end_city_start '))]
assert len(rows)==336,len(rows)
atomic_write(HERE/'structure_end_city_starts_1.21.1.txt', ('\n'.join(rows)+'\n').encode())
print(f'published{len(rows)} actual End City layouts/starts')
