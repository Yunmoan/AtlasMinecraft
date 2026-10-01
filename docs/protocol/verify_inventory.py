"""Read a saved player inventory with the development-only vanilla JAR."""
import argparse
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import zipfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
parser = argparse.ArgumentParser()
parser.add_argument("player_file", type=Path)
args = parser.parse_args()
with tempfile.TemporaryDirectory(prefix="atlas-inventory-reference-") as temp:
    work = Path(temp)
    classpath = []
    with zipfile.ZipFile(ROOT / "jar/server.jar") as jar:
        for name in jar.namelist():
            if name.endswith(".jar") and name.startswith(("META-INF/libraries/", "META-INF/versions/")):
                dest = work / f"library-{len(classpath)}.jar"
                dest.write_bytes(jar.read(name))
                classpath.append(str(dest))
    cp = os.pathsep.join(classpath)
    subprocess.run([shutil.which("javac"), "-proc:none", "-cp", cp, "-d", temp, str(HERE / "PlayerInventoryProbe.java")], check=True, cwd=work)
    subprocess.run([shutil.which("java"), "-cp", os.pathsep.join([temp, cp]), "atlasprobe.PlayerInventoryProbe", str(args.player_file.resolve())], check=True, cwd=work)
