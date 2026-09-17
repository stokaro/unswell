#!/usr/bin/env python3
"""Replay both profiles on frozen sources through an explicitly supplied CLI."""

import argparse
import gzip
from pathlib import Path
import subprocess

from evaluate import archive_sources, read, require, sha, write


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    parser.add_argument('--record', type=Path, default=Path(__file__).resolve().parent.parent)
    parser.add_argument('--compare-reference', action='store_true')
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    output = args.output.resolve()
    require(not output.exists(), 'Output must be new')
    frozen = read(args.record / 'freeze.json')
    for name, digest in frozen['files'].items():
        require(sha((args.record / name).read_bytes()) == digest, 'Frozen artifact changed')
    files = archive_sources(args.record / 'inputs.tar.gz')
    output.mkdir(parents=True)
    inputs = output / 'inputs'
    for name, data in files.items():
        path = inputs / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)
    runs = {}
    reference = read(args.record / 'runs.json')
    for profile in ['technical', 'strict']:
        config = (args.record / (profile + '.yaml')).read_bytes()
        require(sha(config) == reference[profile]['config_sha256'], 'Config drift')
        (inputs / (profile + '.yaml')).write_bytes(config)
        report_path = output / (profile + '.json')
        command = [str(binary), 'check', '--config', profile + '.yaml', '--project-root', '.',
                   '--include-source', '--report', 'json:' + str(report_path), '--timeout', '5m', 'sources']
        result = subprocess.run(command, cwd=inputs, capture_output=True, timeout=360)
        (output / (profile + '.log')).write_bytes(result.stdout + result.stderr)
        require(result.returncode in [0, 1], f'{profile}: operational exit {result.returncode}; see log')
        report = read(report_path)
        require(report['status'] == 'complete' and report['manifest']['complete'] and not report['errors'], 'Incomplete replay')
        compressed = gzip.compress(report_path.read_bytes(), mtime=0)
        name = profile + '.json.gz'
        (output / name).write_bytes(compressed)
        runs[profile] = dict(path=name, sha256=sha(compressed), exit_code=result.returncode,
                             config_sha256=sha(config), binary_sha256=sha(binary.read_bytes()),
                             config_identity_sha256=report['manifest']['config_sources'][0]['sha256'],
                             command=['unswell', *command[1:]])
        if args.compare_reference:
            require(report == read(args.record / name), profile + ': replay differs from frozen report')
        print(f'{profile}: {len(report["documents"])} complete documents; {len(report["findings"])} findings; exit {result.returncode}')
    write(output / 'runs.json', runs)


if __name__ == '__main__':
    main()
