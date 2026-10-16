"""Extract/check the unchanged 1.21.1 End City templates."""
import argparse
import io
import json
from pathlib import Path
import zipfile

ROOT=Path(__file__).resolve().parents[2]
OUTPUT=ROOT/'internal/worldgen/end_city_data'

def main():
    parser=argparse.ArgumentParser();parser.add_argument('--check',action='store_true');args=parser.parse_args()
    with zipfile.ZipFile(ROOT/'jar/server.jar') as bundle:
        version=json.loads(bundle.read('version.json'))
        assert version['id']=='1.21.1' and version['world_version']==3955
        inner=bundle.read('META-INF/versions/1.21.1/server-1.21.1.jar')
    with zipfile.ZipFile(io.BytesIO(inner)) as jar:
        prefix='data/minecraft/structure/end_city/'
        expected={p[len(prefix):]:jar.read(p) for p in jar.namelist() if p.startswith(prefix) and p.endswith('.nbt')}
    assert len(expected)==20
    if args.check:
        assert {p.name for p in OUTPUT.iterdir()}==set(expected)
        for name,data in expected.items():assert (OUTPUT/name).read_bytes()==data,name
        print('verified20 original End City templates')
    else:
        OUTPUT.mkdir(parents=True,exist_ok=True)
        for name,data in expected.items():(OUTPUT/name).write_bytes(data)
        print('extracted20 original End City templates')

if __name__=='__main__':main()
