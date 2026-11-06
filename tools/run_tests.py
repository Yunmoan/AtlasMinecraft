"""Run Atlas test profiles with reference-input preflight and isolated worlds."""
import argparse
from collections import Counter, defaultdict
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

from atlas_tools import ROOT


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--suite', choices=('basic', 'integration', 'reference', 'smoke', 'performance', 'all'), default='basic')
    parser.add_argument('--list', action='store_true', help='show selected tests and missing inputs without running them')
    parser.add_argument('--match', default='', help='only test names containing this text')
    parser.add_argument('--timeout', default='', help='Go timeout per package, e.g. 30m (default: 10m basic, 30m other profiles)')
    args = parser.parse_args()
    go = shutil.which('go')
    if not go:
        parser.error('Go 1.23+ is required; run python tools/check_env.py')
    temporary = ROOT / 'artifacts/.tmp/tests'
    temporary.mkdir(parents=True, exist_ok=True)
    os.environ.update(TEMP=str(temporary), TMP=str(temporary), TMPDIR=str(temporary), GOTMPDIR=str(temporary))
    os.environ.setdefault('GOCACHE', str(ROOT / 'artifacts/go-cache'))
    plan = subprocess.run([go, 'run', './tools/testplan', '-root', str(ROOT)], cwd=ROOT,
                          check=True, capture_output=True, text=True)
    tests = json.loads(plan.stdout)
    print('Available:', ', '.join(f'{suite}={count}' for suite, count in sorted(Counter(t['suite'] for t in tests).items())), flush=True)
    selected = [test for test in tests if (args.suite == 'all' or test['suite'] == args.suite)
                and args.match in test['name']]
    if not selected:
        parser.error('No tests match the selected profile/name.')
    missing = sorted({path for test in selected for path in test.get('inputs', []) if not (ROOT / path).exists()})
    missing_env = sorted({name for test in selected for name in test.get('env', []) if not os.environ.get(name)})
    if args.list:
        for test in selected:
            print(f"{test['suite']:11} {test['package']} {test['name']}")
    if missing or missing_env:
        print('Missing reference inputs (tests have not started):', file=sys.stderr)
        for path in missing:
            print('  ' + path, file=sys.stderr)
        for name in missing_env:
            print('  environment: ' + name, file=sys.stderr)
        print('Generate the matching fixtures using tools/worldgen or tools/protocol; see tools/README.md.', file=sys.stderr)
        return 0 if args.list else 2
    if args.list:
        return 0
    if args.suite in ('basic', 'all') and not args.match:
        subprocess.run([sys.executable, '-m', 'unittest', 'discover', '-s', 'tools', '-p', 'test_*.py'], cwd=ROOT, check=True)
        subprocess.run([go, 'test', './tools/testplan', '-count=1'], cwd=ROOT, check=True)
    if any(test['suite'] in ('smoke', 'performance') for test in selected):
        binaries = ROOT / 'artifacts/test-bin'
        binaries.mkdir(parents=True, exist_ok=True)
        suffix = '.exe' if os.name == 'nt' else ''
        for target in ('atlas', 'atlas-test'):
            subprocess.run([go, 'build', '-o', str(binaries / (target + suffix)), './cmd/' + target], cwd=ROOT, check=True)
        os.environ['ATLAS_SERVER_EXE'] = str(binaries / ('atlas' + suffix))
        os.environ['ATLAS_TEST_EXE'] = str(binaries / ('atlas-test' + suffix))
        os.environ['ATLAS_SMOKE_DIR'] = str(ROOT / 'artifacts/test-runs')
    else:
        # A basic/reference run must not opt in to binary scenarios inherited
        # from the shell or start a Java interoperability server accidentally.
        for name in ('ATLAS_SERVER_EXE', 'ATLAS_TEST_EXE'):
            os.environ.pop(name, None)
    groups = defaultdict(list)
    for test in selected:
        groups[test['package']].append(test['name'])
    for package, names in sorted(groups.items()):
        print(f'Running {len(names)} {args.suite} tests in {package}', flush=True)
        timeout = args.timeout or ('10m' if args.suite == 'basic' else '30m')
        result = subprocess.run([go, 'test', package, '-run', '^(' + '|'.join(names) + ')$', '-count=1', '-timeout=' + timeout], cwd=ROOT)
        if result.returncode:
            return result.returncode
    print(f'Passed {len(selected)} selected tests.', flush=True)
    return 0


if __name__ == '__main__':
    try:
        raise SystemExit(main())
    except subprocess.CalledProcessError as error:
        print(error.stderr or str(error), file=sys.stderr)
        raise SystemExit(error.returncode)
