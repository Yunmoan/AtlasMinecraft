"""Read Java 1.21.1 entity IDs from the supplied development reference JAR."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import zipfile
import sys

from _bootstrap import common

HERE = Path(__file__).resolve().parent
PROBE_DIR = HERE / "probes"
ROOT = common.ROOT
with tempfile.TemporaryDirectory(prefix="atlas-player-reference-") as temp:
    work = Path(temp)
    classpath = [str(path) for path in common.reference_jars(work)]
    common.java_tools()
    cp = os.pathsep.join(classpath)
    subprocess.run([shutil.which("javac"), "-proc:none", "-cp", cp, "-d", temp, str(PROBE_DIR / "PlayerEntityProbe.java")], check=True, cwd=work)
    subprocess.run([shutil.which("java"), "-cp", os.pathsep.join([temp, cp]), "atlasprobe.PlayerEntityProbe", *[str(Path(arg).resolve()) for arg in sys.argv[1:]]], check=True, cwd=work)
