import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

import atlas_tools as common


class ReferenceToolsTest(unittest.TestCase):
    def tearDown(self):
        common.java_tools.cache_clear()
        common.reference_jar.cache_clear()

    def test_java_home_selects_both_tools_and_updates_path(self):
        with tempfile.TemporaryDirectory() as temporary:
            directory = Path(temporary) / 'bin'
            directory.mkdir()
            for name in ('java', 'javac'):
                (directory / (name + ('.exe' if os.name == 'nt' else ''))).touch()
            completed = subprocess.CompletedProcess([], 0, 'javac 21.0.1', 'java version "21.0.1"')
            with patch.dict(os.environ, {'JAVA_HOME': temporary}), patch('subprocess.run', return_value=completed) as run:
                java, javac = common.java_tools()
                self.assertEqual(Path(java).parent, directory)
                self.assertEqual(Path(javac).parent, directory)
                self.assertEqual(run.call_count, 2)
                self.assertTrue(os.environ['PATH'].startswith(str(directory) + os.pathsep))

    def test_incomplete_java_home_is_an_error(self):
        with tempfile.TemporaryDirectory() as temporary, patch.dict(os.environ, {'JAVA_HOME': temporary}):
            with self.assertRaisesRegex(RuntimeError, 'complete JDK'):
                common.java_tools()

    def test_old_jdk_is_rejected(self):
        completed = subprocess.CompletedProcess([], 0, '', 'java version "17.0.1"')
        with patch.dict(os.environ, {'JAVA_HOME': ''}), patch('shutil.which', return_value='/jdk/bin/java'), patch('subprocess.run', return_value=completed):
            with self.assertRaisesRegex(RuntimeError, 'JDK 21'):
                common.java_tools()

    def test_reference_missing_or_wrong_hash_is_rejected(self):
        with tempfile.TemporaryDirectory() as temporary, patch.object(common, 'ROOT', Path(temporary)):
            with self.assertRaisesRegex(RuntimeError, 'Missing'):
                common.reference_jar()
            path = Path(temporary) / 'jar/server.jar'
            path.parent.mkdir()
            path.write_bytes(b'wrong version')
            with self.assertRaisesRegex(RuntimeError, 'SHA-1'):
                common.reference_jar()

    def test_atomic_write_replaces_file_and_creates_parent(self):
        with tempfile.TemporaryDirectory() as temporary:
            path = Path(temporary) / 'nested/result.bin'
            common.atomic_write(path, b'first')
            common.atomic_write(path, b'second')
            self.assertEqual(path.read_bytes(), b'second')
            self.assertEqual(list(path.parent.glob('*.tmp')), [])

    def test_atomic_write_retries_sharing_violation(self):
        replace = os.replace
        with tempfile.TemporaryDirectory() as temporary:
            path = Path(temporary) / 'result.bin'
            calls = 0
            def retry(source, target):
                nonlocal calls
                calls += 1
                if calls == 1:
                    raise PermissionError()
                return replace(source, target)
            with patch('os.replace', side_effect=retry) as swap, patch('time.sleep'):
                common.atomic_write(path, b'complete')
                self.assertEqual(swap.call_count, 2)
            self.assertEqual(path.read_bytes(), b'complete')


if __name__ == '__main__':
    unittest.main()
