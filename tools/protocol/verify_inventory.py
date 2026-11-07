"""Read a saved player inventory with the development-only vanilla JAR."""
import argparse
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import zipfile

from _bootstrap import common

HERE = Path(__file__).resolve().parent
PROBE_DIR = HERE / "probes"
ROOT = common.ROOT
parser = argparse.ArgumentParser()
parser.add_argument("player_file", type=Path)
args = parser.parse_args()
with tempfile.TemporaryDirectory(prefix="atlas-inventory-reference-") as temp:
    work = Path(temp)
    classpath = [str(path) for path in common.reference_jars(work)]
    common.java_tools()
    cp = os.pathsep.join(classpath)
    subprocess.run([shutil.which("javac"), "-proc:none", "-cp", cp, "-d", temp, str(PROBE_DIR / "PlayerInventoryProbe.java")], check=True, cwd=work)
    subprocess.run([shutil.which("java"), "-cp", os.pathsep.join([temp, cp]), "atlasprobe.PlayerInventoryProbe", str(args.player_file.resolve())], check=True, cwd=work)
