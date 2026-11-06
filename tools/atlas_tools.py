"""Shared paths and official-reference tooling; standard library only."""
from functools import lru_cache
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import tempfile
import time
import zipfile

ROOT = Path(__file__).resolve().parents[1]
REFERENCE_VERSION = '1.21.1'
REFERENCE_DATA_VERSION = 3955
REFERENCE_SHA1 = '59353fb40c36d304f2035d51e7d6e6baa98dc05c'


@lru_cache(maxsize=1)
def java_tools():
    """Return (java, javac), preferring one complete JDK from JAVA_HOME."""
    suffix = '.exe' if os.name == 'nt' else ''
    home = os.environ.get('JAVA_HOME')
    if home:
        directory = Path(home) / 'bin'
        java, javac = (str(directory / (name + suffix)) for name in ('java', 'javac'))
        if not all(Path(path).is_file() for path in (java, javac)):
            raise RuntimeError('JAVA_HOME must point to a complete JDK (java and javac).')
    else:
        java, javac = shutil.which('java'), shutil.which('javac')
        if not java or not javac:
            raise RuntimeError('JDK 21+ is required; set JAVA_HOME or add java and javac to PATH.')
        directory = Path(java).parent
    for executable in (java, javac):
        try:
            result = subprocess.run([executable, '-version'], capture_output=True, text=True, timeout=10, check=True)
        except (OSError, subprocess.SubprocessError) as error:
            raise RuntimeError('Cannot run the JDK; set JAVA_HOME to its installation directory.') from error
        version = re.search(r'(?:version\s+"|javac\s+)(\d+)', result.stdout + result.stderr)
        if not version or int(version[1]) < 21:
            raise RuntimeError('The reference probes require JDK 21 or newer.')
    # Legacy probe commands use java/javac by name. Keep both on this JDK.
    os.environ['PATH'] = str(directory) + os.pathsep + os.environ.get('PATH', '')
    return java, javac


@lru_cache(maxsize=1)
def reference_jar():
    """Locate and validate the pinned official bundle before extracting data."""
    path = ROOT / 'jar/server.jar'
    if not path.is_file():
        raise RuntimeError('Missing jar/server.jar: supply the official Minecraft Java 1.21.1 server bundle.')
    with path.open('rb') as stream:
        actual = hashlib.file_digest(stream, 'sha1').hexdigest()
    if actual != REFERENCE_SHA1:
        raise RuntimeError('jar/server.jar does not match the pinned official Minecraft 1.21.1 SHA-1.')
    with zipfile.ZipFile(path) as archive:
        version = json.loads(archive.read('version.json'))
    if version.get('id') != REFERENCE_VERSION or version.get('world_version') != REFERENCE_DATA_VERSION:
        raise RuntimeError('Expected Minecraft Java 1.21.1 / DataVersion 3955.')
    return path


def reference_jars(destination):
    """Extract the validated bundle and return [inner server, libraries...]."""
    destination = Path(destination)
    destination.mkdir(parents=True, exist_ok=True)
    with zipfile.ZipFile(reference_jar()) as archive:
        members = ['META-INF/versions/1.21.1/server-1.21.1.jar']
        members += sorted(name for name in archive.namelist()
                          if name.startswith('META-INF/libraries/') and name.endswith('.jar'))
        paths = []
        for index, member in enumerate(members):
            path = destination / ('server-1.21.1.jar' if index == 0 else f'library-{index}.jar')
            atomic_write(path, archive.read(member))
            paths.append(path)
    return paths


@lru_cache(maxsize=1)
def _prepared_reference():
    cache = ROOT / 'artifacts/reference-java'
    stamp = cache / 'bundle.json'
    if stamp.is_file():
        manifest = json.loads(stamp.read_text())
        paths = [cache / name for name in manifest.get('files', [])]
        if manifest.get('sha1') == REFERENCE_SHA1 and len(paths) > 1 and all(path.is_file() for path in paths):
            reference_jar()
            return paths
    paths = reference_jars(cache)
    atomic_write(stamp, json.dumps({'sha1': REFERENCE_SHA1, 'files': [path.name for path in paths]}).encode())
    return paths


def reference_classpath(classes):
    """Prepare one checked reference cache and include the caller's classes."""
    java_tools()
    classes = Path(classes)
    classes.mkdir(parents=True, exist_ok=True)
    return os.pathsep.join(map(str, [classes, *_prepared_reference()]))


def reference_server():
    """Return the validated inner server JAR for data readers and probes."""
    return _prepared_reference()[0]


def atomic_write(path, encoded):
    """Publish complete output, retrying Windows sharing violations."""
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    descriptor, name = tempfile.mkstemp(prefix=path.name + '.', suffix='.tmp', dir=path.parent)
    temporary = Path(name)
    published = False
    try:
        with os.fdopen(descriptor, 'wb') as stream:
            stream.write(encoded)
            stream.flush()
            os.fsync(stream.fileno())
        for attempt in range(60):
            try:
                os.replace(temporary, path)
                published = True
                break
            except PermissionError:
                if attempt == 59:
                    raise PermissionError(f'Output is still open; unpublished data retained at {temporary}')
                time.sleep(0.2)
    finally:
        if published:
            temporary.unlink(missing_ok=True)
