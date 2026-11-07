"""Check build prerequisites, optionally including official-reference tools."""
import argparse
import re
import shutil
import subprocess
import sys

import atlas_tools as common


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--reference', action='store_true', help='also check JDK 21+ and the pinned server JAR')
    args = parser.parse_args()
    try:
        go = shutil.which('go')
        if not go:
            raise RuntimeError('Go 1.23+ is required; add go to PATH.')
        result = subprocess.run([go, 'version'], capture_output=True, text=True, check=True, timeout=10)
        version = re.search(r'go(\d+)\.(\d+)', result.stdout)
        if not version or (int(version[1]), int(version[2])) < (1, 23):
            raise RuntimeError('Go 1.23 or newer is required.')
        print(result.stdout.strip())
        if args.reference:
            java, javac = common.java_tools()
            print('java:', java)
            print('javac:', javac)
            print('reference:', common.reference_jar())
        print('Environment checks passed.')
        return 0
    except (RuntimeError, OSError, subprocess.SubprocessError) as error:
        print(error, file=sys.stderr)
        return 1


if __name__ == '__main__':
    raise SystemExit(main())
