#!/usr/bin/env python3
"""Replay 36 complete sources with an explicit local binary and fresh outputs."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import tarfile

ROOT = Path(__file__).resolve().parent
REFERENCES = {
    'whole': ROOT.parent / '2026-09-17-full-page-recall',
    'context': ROOT.parent / '2026-09-17-proposition-repetition/confirmation',
}


def read(path):
    return json.loads(path.read_text())


def sources(root=ROOT):
    result = {}
    for cohort, reference in REFERENCES.items():
        with tarfile.open(reference / 'inputs.tar.gz') as archive:
            for page in read(reference / 'manifest.json')['pages']:
                data = archive.extractfile(page['path']).read()
                expected = page.get('sha256', page.get('source_sha256'))
                if hashlib.sha256(data).hexdigest() != expected:
                    raise ValueError('Reference source drift')
                result[cohort + '/' + page['id'] + Path(page['path']).suffix] = data
    for page in read(root / 'confirmation-inputs.json')['pages']:
        data = page['source'].encode()
        if hashlib.sha256(data).hexdigest() != page['sha256']:
            raise ValueError('Confirmation source drift')
        result['confirmation/' + page['id'] + Path(page['path']).suffix] = data
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    binary = args.binary.resolve(strict=True)
    output = args.output.resolve()
    output.mkdir(parents=True, exist_ok=False)
    inputs = output / 'inputs'
    for name, data in sources().items():
        path = inputs / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)
    runs = {}
    for profile in ['technical', 'strict']:
        policy = f'version: 1\nextends: [builtin:{profile}-v1]\n'
        (inputs / 'policy.yaml').write_text(policy)
        report = output / (profile + '.json')
        command = [str(binary), 'check', '--config', 'policy.yaml', '--project-root', '.',
                   '--include-source', '--timeout', '5m', '--report', 'json:' + str(report),
                   'whole', 'context', 'confirmation']
        run = subprocess.run(command, cwd=inputs, capture_output=True, timeout=360)
        (output / (profile + '.log')).write_bytes(run.stdout + run.stderr)
        if run.returncode not in (0, 1):
            raise ValueError(f'Operational exit {run.returncode}; see {profile}.log')
        data = read(report)
        if data['status'] != 'complete' or not data['manifest']['complete'] or data['errors']:
            raise ValueError('Incomplete report')
        packed = gzip.compress((json.dumps(data, separators=(',', ':')) + '\n').encode(), mtime=0)
        report.with_suffix('.json.gz').write_bytes(packed)
        runs[profile] = dict(binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                             report_sha256=hashlib.sha256(packed).hexdigest(), exit_code=run.returncode,
                             policy_sha256=hashlib.sha256(policy.encode()).hexdigest(), command=command)
        print(profile, len(data['documents']), 'documents', len(data['findings']), 'findings', flush=True)
    (output / 'runs.json').write_text(json.dumps(runs, indent=2) + '\n')


if __name__ == '__main__':
    main()
