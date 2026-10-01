"""Extract official classes for source-level worldgen port review."""
import subprocess
import zipfile
from pathlib import Path

root = Path(__file__).resolve().parents[2]
target = root / "artifacts/worldgen-probes/remaining-source"
target.mkdir(exist_ok=True)
with zipfile.ZipFile(root / "vanilla/server-inner.jar") as jar:
    for name in ["edp", "edk", "ecs", "edj", "ebh", "ecq", "edm", "edn", "ect", "egv", "egy", "egz", "ehc", "eha$a", "ehj", "ehk", "ehg", "ehi", "efw", "efx", "efs", "ebb", "eaw", "ebe", "dzb", "def", "ddw", "dni", "dmx", "dkc", "dkd", "dmw", "dfc", "eib", "ehv", "ehz", "eby", "ebz", "ecu", "eda", "eds", "ecw", "edm"]:
        if (target / f"{name}.java").exists():
            continue
        path = target / f"{name}.class"
        path.write_bytes(jar.read(f"{name}.class"))
        subprocess.run(["java", "-jar", str(root / "artifacts/worldgen-probes/cfr.jar"), str(path), "--outputdir", str(target), "--silent", "true"], check=True, capture_output=True)
print(f"Extracted sources to {target}")
