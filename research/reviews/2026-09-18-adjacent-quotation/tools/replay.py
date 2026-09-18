#!/usr/bin/env python3
"""Compare the quotation repair with the saved main reports on exposed pages."""
import argparse
import gzip
import hashlib
import json
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
STUDY = ROOT.parent / '2026-09-18-discourse-patterns'


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def report(path):
    value = json.loads(gzip.decompress(path.read_bytes()))
    if value['status'] != 'complete' or not value['manifest']['complete'] or value['errors']:
        raise ValueError('Incomplete report: ' + str(path))
    if value.get('abstentions') or value['manifest'].get('skipped_rules'):
        raise ValueError('Unavailable analysis: ' + str(path))
    return value


def key(finding):
    primary = finding['primary']
    return (finding['rule_id'], primary['path'], primary['span']['start'], primary['span']['end'],
            tuple((item['span']['start'], item['span']['end']) for item in finding['related']))


def comparison(dataset, profile, output):
    before_path = dataset/'before'/(profile+'.json.gz')
    after_path = output/dataset.name/(profile+'.json.gz')
    before, after = report(before_path), report(after_path)
    old = {key(item): item for item in before['findings']}
    new = {key(item): item for item in after['findings']}
    added = [new[k] for k in sorted(new.keys()-old.keys())]
    removed = [old[k] for k in sorted(old.keys()-new.keys())]
    if any(item['rule_id'] != 'repetition.adjacent-word' for item in added+removed):
        raise ValueError('Another rule changed')
    return dict(dataset=dataset.name, profile=profile, documents=len(after['documents']),
                before=len(old), after=len(new), before_report_sha256=digest(before_path),
                after_report_sha256=digest(after_path), added=added, removed=removed)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--binary', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    binary, output = args.binary.resolve(strict=True), args.output.resolve()
    if output.exists():
        raise ValueError('Use a new output directory')
    rows = []
    for dataset in sorted((STUDY/'reports').iterdir()):
        subprocess.run([sys.executable, str(STUDY/'tools/measure.py'), '--binary', str(binary),
                        '--set', dataset.name, '--output', str(output/dataset.name)], check=True)
        rows.extend(comparison(dataset, profile, output) for profile in ('technical', 'strict'))
    result = dict(kind='exposed-regression', binary_sha256=digest(binary), rows=rows)
    (output/'regression.json').write_text(json.dumps(result, indent=2)+'\n')


if __name__ == '__main__':
    main()
