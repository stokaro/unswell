#!/usr/bin/env python3
"""Verify paired locations, unchanged rules, and the retained confirmation miss."""
import collections
import gzip
import json
from pathlib import Path
import sys
sys.dont_write_bytecode = True
from measure import ROOT, sha, sources

RULE = 'filler.numbered-section-framing'


def locations(finding):
    return [finding['primary'], *finding['related']]


def identity(finding):
    return json.dumps({k: finding[k] for k in ('rule_id', 'rule_version', 'message', 'primary', 'related', 'evidence')}, sort_keys=True)


def check_pair(before, after, review):
    for report in (before, after):
        if report['status'] != 'complete' or report['errors'] or report.get('abstentions') or report['manifest']['skipped_rules']:
            raise ValueError('Incomplete analysis')
    new = [f for f in after['findings'] if f['rule_id'] == RULE]
    old = collections.Counter(identity(f) for f in before['findings'])
    kept = collections.Counter(identity(f) for f in after['findings'] if f['rule_id'] != RULE)
    if old != kept:
        raise ValueError('An existing diagnostic changed')
    if len(new) != 1:
        raise ValueError('Expected exactly one reviewed new finding')
    wanted = review['development'][0]['targets']
    actual = locations(new[0])
    if len(actual) != len(wanted):
        raise ValueError('Missing paired source location')
    for location, target in zip(actual, wanted):
        expected_end = target['end'] - int(target['quote'].endswith('.'))
        if location['path'] != target['path'] or location['span'] != {'start': target['start'], 'end': expected_end}:
            raise ValueError('Wrong source location; nearby metadiscourse is not coverage')
    missed = review['confirmation'][0]['targets'][0]
    for finding in after['findings']:
        for location in locations(finding):
            if location['path'] == missed['path'] and location['span']['start'] < missed['end'] and location['span']['end'] > missed['start']:
                raise ValueError('Confirmation miss needs an explicit review update')
    return {'baseline_findings': len(before['findings']), 'candidate_findings': len(after['findings']),
            'new_actionable': 1, 'new_confirmation_findings': 0, 'confirmation_target_defects': 1, 'confirmation_detected': 0}


def evaluate():
    data = sources()
    review = json.loads((ROOT/'source-review.json').read_text())
    for event in review['development']+review['confirmation']:
        for target in event['targets']:
            if data[target['path']][target['start']:target['end']].decode() != target['quote']:
                raise ValueError('Annotation drift')
    freeze = json.loads((ROOT/'code-freeze.json').read_text())
    repository = ROOT.parents[2]
    for name, digest in freeze['runtime_changes'].items():
        if sha((repository/name).read_bytes()) != digest:
            raise ValueError('Frozen candidate changed')
    result = {}
    for profile in ('technical', 'strict'):
        reports = []
        for phase in ('before', 'after'):
            path = ROOT/'reports'/phase
            record = json.loads((path/'runs.json').read_text())[profile]
            packed = (path/(profile+'.json.gz')).read_bytes()
            if sha(packed) != record['report_sha256']:
                raise ValueError('Report hash mismatch')
            report = json.loads(gzip.decompress(packed))
            key = 'baseline_commit' if phase == 'before' else 'candidate_commit'
            if report['manifest']['tool_commit'] != freeze[key]:
                raise ValueError('Wrong runtime identity')
            if phase == 'after' and record['binary_sha256'] != freeze['binary_sha256']:
                raise ValueError('Wrong frozen binary')
            if {d['name'] for d in report['documents']} != {p for p in data if p.startswith(('sources/', 'development/'))}:
                raise ValueError('Missing page')
            reports.append(report)
        result[profile] = check_pair(*reports, review)
    return result


if __name__ == '__main__':
    print(json.dumps(evaluate(), indent=2))
