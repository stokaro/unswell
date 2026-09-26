#!/usr/bin/env python3
"""Prepare and validate a source-bound, label-free contextual review experiment."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
BASE = ROOT.parent / '2026-09-26-outcome-announcements'
CATEGORIES = ['empty_framing', 'needless_repetition', 'wordiness', 'vague_claims',
              'needless_complexity', 'unsupported_evaluation', 'other_editorial']


def digest(data):
    return hashlib.sha256(data).hexdigest()


def read(path):
    return json.loads(path.read_text())


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + '\n')


def require(condition, message):
    if not condition:
        raise ValueError(message)


def object_schema(properties):
    return {'type': 'object', 'additionalProperties': False,
            'properties': properties, 'required': list(properties)}


def schema():
    string = {'type': 'string'}
    target = object_schema({'unit': string, 'quote': string,
                            'occurrence': {'type': 'integer', 'minimum': 1}})
    finding = object_schema({'id': string, 'category': {'enum': CATEGORIES, 'type': 'string'},
                             'targets': {'type': 'array', 'items': target},
                             'diagnostic': string, 'why_revision_needed': string,
                             'suggestion': string, 'meaning_preservation': string})
    coverage = object_schema({'unit': string,
                              'status': {'enum': ['reviewed', 'uncertain'], 'type': 'string'}})
    return object_schema({'page': string, 'coverage': {'type': 'array', 'items': coverage},
                           'findings': {'type': 'array', 'items': finding}})


def prepare(output):
    packed = (BASE / 'after-technical.json.gz').read_bytes()
    report = json.loads(gzip.decompress(packed))
    require(report['status'] == 'complete' and report['manifest']['complete'] and
            not report['errors'], 'Incomplete baseline')
    output.mkdir(parents=True, exist_ok=False)
    (output / 'requests').mkdir()
    for name in ['prompt.md', 'study.py', 'run.py', 'protocol.json']:
        (output / name).write_bytes((ROOT / name).read_bytes())
    write(output / 'response-schema.json', schema())
    records = []
    for ordinal, doc in enumerate(sorted(report['documents'], key=lambda d: d['name'])):
        data = doc['source'].encode()
        require(digest(data) == doc['source_hash'], 'Source identity changed')
        units = []
        for a in report['assessments']:
            if a['path'] != doc['name'] or a['scope'] != 'paragraph':
                continue
            require(a['status'] == 'available', 'Unavailable baseline unit')
            start, end = a['span']['start'], a['span']['end']
            require(0 <= start < end <= len(data), 'Invalid unit range')
            data[start:end].decode()
            units.append({'id': f"u{a['unit_id']:04}", 'span': a['span']})
        require(len({u['id'] for u in units}) == len(units), 'Duplicate units')
        page = f'page-{ordinal + 1:03}'
        packet = {'page': page, 'source_sha256': digest(data), 'source': doc['source'],
                  'format': doc['format'], 'units': units, 'protected': doc['excluded']}
        target = output / 'requests' / (page + '.json')
        write(target, packet)
        records.append({'page': page, 'reference': doc['name'],
                        'request_sha256': digest(target.read_bytes()),
                        'source_sha256': digest(data), 'bytes': len(data), 'units': len(units)})
    require(len(records) == 36, 'Frozen full-page scope changed')
    write(output / 'manifest.json', {'version': 1, 'phase': 'exposed_development',
                                     'baseline_report_sha256': digest(packed),
                                     'source_tree': '82a4e0023f8f36cc73a521ac871e93a5d204b150',
                                     'pages': records})
    freeze(output)


def freeze(output):
    files = sorted(p for p in output.rglob('*') if p.is_file() and p.name != 'freeze.json')
    write(output / 'freeze.json', {str(p.relative_to(output)): digest(p.read_bytes()) for p in files})


def verify_freeze(output):
    hashes = read(output / 'freeze.json')
    actual = {str(p.relative_to(output)) for p in output.rglob('*')
              if p.is_file() and p.name != 'freeze.json'}
    require(set(hashes) == actual, 'Frozen file set changed')
    for name, expected in hashes.items():
        require(digest((output / name).read_bytes()) == expected, 'Frozen input changed: ' + name)
    return hashes


def exact_keys(value, keys):
    require(isinstance(value, dict) and set(value) == set(keys), 'Unexpected object fields')


def bind_target(packet, target):
    exact_keys(target, ['unit', 'quote', 'occurrence'])
    units = {u['id']: u for u in packet['units']}
    require(target['unit'] in units, 'Unknown unit')
    require(type(target['occurrence']) is int and target['occurrence'] >= 1, 'Invalid occurrence')
    require(isinstance(target['quote'], str) and target['quote'].strip(), 'Empty target')
    data = packet['source'].encode()
    unit = units[target['unit']]['span']
    haystack = data[unit['start']:unit['end']]
    needle = target['quote'].encode()
    at = -len(needle)
    for _ in range(target['occurrence']):
        at = haystack.find(needle, at + len(needle))
        require(at >= 0, 'Quotation is not in the selected source unit')
    start, end = unit['start'] + at, unit['start'] + at + len(needle)
    eligible = bytearray(data[start:end])
    for protected in packet['protected']:
        span = protected['span']
        left, right = max(start, span['start']), min(end, span['end'])
        if left < right:
            eligible[left-start:right-start] = b' ' * (right-left)
    require(any(chr(c).isalnum() for c in eligible if c < 128), 'Target contains no eligible prose')
    return {'span': {'start': start, 'end': end}, 'quote': target['quote'], 'unit': target['unit']}


def validate(packet, response):
    require(digest(packet['source'].encode()) == packet['source_sha256'], 'Packet source drift')
    exact_keys(response, ['page', 'coverage', 'findings'])
    require(response['page'] == packet['page'], 'Wrong page')
    require(isinstance(response['coverage'], list) and isinstance(response['findings'], list),
            'Expected lists')
    seen = set()
    for row in response['coverage']:
        exact_keys(row, ['unit', 'status'])
        require(row['unit'] not in seen, 'Duplicate coverage')
        require(row['status'] in ['reviewed', 'uncertain'], 'Unknown coverage status')
        seen.add(row['unit'])
    require(seen == {u['id'] for u in packet['units']}, 'Incomplete or invented coverage')
    found = set()
    bound = []
    require(len(response['findings']) <= 128, 'Finding limit exceeded')
    for finding in response['findings']:
        exact_keys(finding, ['id', 'category', 'targets', 'diagnostic', 'why_revision_needed',
                             'suggestion', 'meaning_preservation'])
        require(isinstance(finding['id'], str) and finding['id'].strip() and
                finding['id'] not in found, 'Missing or duplicate finding ID')
        found.add(finding['id'])
        require(finding['category'] in CATEGORIES, 'Unknown category')
        for field in ['diagnostic', 'why_revision_needed', 'suggestion', 'meaning_preservation']:
            require(isinstance(finding[field], str) and finding[field].strip(), 'Missing explanation')
        require(isinstance(finding['targets'], list) and 1 <= len(finding['targets']) <= 8,
                'Invalid target count')
        targets = [bind_target(packet, t) for t in finding['targets']]
        require(len({(t['span']['start'], t['span']['end']) for t in targets}) == len(targets),
                'Duplicate target')
        bound.append({**finding, 'targets': targets})
    return {'page': packet['page'], 'source_sha256': packet['source_sha256'],
            'coverage': response['coverage'], 'findings': bound}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest='command', required=True)
    prep = sub.add_parser('prepare'); prep.add_argument('output', type=Path)
    check = sub.add_parser('verify'); check.add_argument('output', type=Path)
    args = parser.parse_args()
    if args.command == 'prepare':
        prepare(args.output)
    else:
        hashes = verify_freeze(args.output)
        print(json.dumps({'verified_files': len(hashes)}))


if __name__ == '__main__':
    main()
