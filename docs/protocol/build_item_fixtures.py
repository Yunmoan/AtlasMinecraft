"""Generate official wire samples and built-in potion/effect registry IDs."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import zipfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
DEST = ROOT / "internal/item"
with tempfile.TemporaryDirectory(prefix="atlas-item-wire-") as temp:
    work = Path(temp)
    classpath = []
    with zipfile.ZipFile(ROOT / "jar/server.jar") as jar:
        for name in jar.namelist():
            if name.endswith(".jar") and name.startswith(("META-INF/libraries/", "META-INF/versions/")):
                dest = work / f"library-{len(classpath)}.jar"
                dest.write_bytes(jar.read(name))
                classpath.append(str(dest))
    cp = os.pathsep.join(classpath)
    subprocess.run([shutil.which("javac"), "-proc:none", "-cp", cp, "-d", temp, str(HERE / "ItemStackProbe.java")], check=True, cwd=work)
    subprocess.run([shutil.which("java"), "-cp", os.pathsep.join([temp, cp]), "atlasprobe.ItemStackProbe",
                    str(DEST / "testdata/vanilla_stacks_767.json"), str(DEST / "builtin_registries_767.json")], check=True, cwd=work)
