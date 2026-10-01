"""Read Java 1.21.1 entity IDs from the supplied development reference JAR."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import zipfile
import sys

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
with tempfile.TemporaryDirectory(prefix="atlas-player-reference-") as temp:
    work = Path(temp)
    classpath = []
    with zipfile.ZipFile(ROOT / "jar/server.jar") as jar:
        for name in jar.namelist():
            if name.endswith(".jar") and name.startswith(("META-INF/libraries/", "META-INF/versions/")):
                dest = work / f"library-{len(classpath)}.jar"
                dest.write_bytes(jar.read(name))
                classpath.append(str(dest))
    cp = os.pathsep.join(classpath)
    subprocess.run([shutil.which("javac"), "-proc:none", "-cp", cp, "-d", temp, str(HERE / "PlayerEntityProbe.java")], check=True, cwd=work)
    subprocess.run([shutil.which("java"), "-cp", os.pathsep.join([temp, cp]), "atlasprobe.PlayerEntityProbe", *[str(Path(arg).resolve()) for arg in sys.argv[1:]]], check=True, cwd=work)
