"""Pin Paper's archived 1.21.1 branch and retain relevant patches for review."""
import json
import io
import urllib.request
import zipfile
from pathlib import Path

target = Path(__file__).resolve().parent / "paper-1.21.1"
target.mkdir(exist_ok=True)

def fetch(url):
    request = urllib.request.Request(url, headers={"User-Agent": "Atlas-worldgen-reference"})
    with urllib.request.urlopen(request, timeout=45) as response:
        return response.read()

repo = "PaperMC/Paper-archive"
info = json.loads(fetch(f"https://api.github.com/repos/{repo}/branches/ver/1.21.1"))
commit = info["commit"]["sha"]
tree = json.loads(fetch(f"https://api.github.com/repos/{repo}/git/trees/{commit}?recursive=1"))
terms = ("worldgen", "world-gen", "noise", "aquifer", "biome", "generation", "heightmap", "cave", "chunk-system")
paths = [entry["path"] for entry in tree["tree"] if entry["type"] == "blob" and entry["path"].endswith(".patch") and any(term in entry["path"].lower() for term in terms)]
(target / "manifest.json").write_text(json.dumps({"repository": repo, "commit": commit, "paths": paths}, indent=2), encoding="utf-8")
archive = fetch(f"https://api.github.com/repos/{repo}/zipball/{commit}")
with zipfile.ZipFile(io.BytesIO(archive)) as source:
    for entry in source.infolist():
        path = entry.filename.split("/", 1)[-1]
        if path.endswith(".patch"):
            content = source.read(entry)
            if path in paths or any(term in content for term in (b"NoiseChunk", b"NoiseBasedChunkGenerator", b"Aquifer", b"DensityFunctions", b"SurfaceSystem", b"Climate.java")):
                (target / Path(path).name).write_bytes(content)
print(f"Paper 1.21.1 commit {commit}: downloaded {len(paths)} relevant patches")
