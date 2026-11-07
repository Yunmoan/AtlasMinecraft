"""Generate official wire samples and built-in potion/effect registry IDs."""
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
DEST = ROOT / "internal/item"
with tempfile.TemporaryDirectory(prefix="atlas-item-wire-") as temp:
    work = Path(temp)
    classpath = [str(path) for path in common.reference_jars(work)]
    common.java_tools()
    cp = os.pathsep.join(classpath)
    subprocess.run([shutil.which("javac"), "-proc:none", "-cp", cp, "-d", temp, str(PROBE_DIR / "ItemStackProbe.java")], check=True, cwd=work)
    subprocess.run([shutil.which("java"), "-cp", os.pathsep.join([temp, cp]), "atlasprobe.ItemStackProbe",
                    str(DEST / "testdata/vanilla_stacks_767.json"), str(DEST / "builtin_registries_767.json")], check=True, cwd=work)
