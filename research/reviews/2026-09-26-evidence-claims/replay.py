#!/usr/bin/env python3
"""Replay all 30 complete pages using an explicit local CLI; never overwrite evidence."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parent
REFERENCE = ROOT.parent / '2026-09-17-full-page-recall'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=False)
    subprocess.run([sys.executable, str(REFERENCE / 'tools/measure.py'), '--binary', str(binary),
                    '--output', str(output / 'development')], check=True)
    inputs = output / 'confirmation-inputs'
    inputs.mkdir()
    pages = json.loads((ROOT / 'confirmation-inputs.json').read_text())['pages']
    for page in pages:
        data = page['source'].encode()
        if hashlib.sha256(data).hexdigest() != page['sha256']:
            raise ValueError('Source hash mismatch')
        (inputs / (page['id'] + '.md')).write_bytes(data)
    runs = {}
    for profile in ['technical', 'strict']:
        (inputs / 'policy.yaml').write_text(f'version: 1\nextends: [builtin:{profile}-v1]\n')
        report = output / ('confirmation-' + profile + '.json')
        command = [str(binary), 'check', '--config', 'policy.yaml', '--project-root', '.', '--include-source',
                   '--timeout', '5m', '--report', 'json:' + str(report), *[p['id'] + '.md' for p in pages]]
        result = subprocess.run(command, cwd=inputs, capture_output=True, timeout=360)
        (output / (profile + '.log')).write_bytes(result.stdout + result.stderr)
        if result.returncode not in (0, 1):
            raise RuntimeError(f'Operational failure: {profile}, exit {result.returncode}')
        data = json.loads(report.read_bytes())
        if data['status'] != 'complete' or not data['manifest']['complete'] or data['errors']:
            raise ValueError('Incomplete confirmation')
        compressed = gzip.compress(report.read_bytes(), mtime=0)
        report.with_suffix('.json.gz').write_bytes(compressed)
        runs[profile] = dict(command=command, exit_code=result.returncode,
                             binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                             report_sha256=hashlib.sha256(compressed).hexdigest())
    (output / 'confirmation-runs.json').write_text(json.dumps(runs, indent=2) + '\n')


if __name__ == '__main__':
    main()
