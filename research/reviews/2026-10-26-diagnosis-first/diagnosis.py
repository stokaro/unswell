#!/usr/bin/env python3
"""Prepare and bind a diagnosis-only pilot from unchanged source units."""
import argparse
import copy
import importlib.util
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
ACCOUNTING = ROOT.parent / '2026-09-27-contextual-accounting'
SPEC = importlib.util.spec_from_file_location('frozen_accounting_source', ACCOUNTING / 'accounting.py')
OLD = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(OLD)
REFERENCE = OLD.REFERENCE
digest, read, write, require = OLD.digest, OLD.read, OLD.write, OLD.require
MAX_UNITS = OLD.MAX_UNITS
PAGES = [f'context/c{i:02}.md' for i in range(1, 7)]


def schema():
    string = {'type': 'string'}
    target = REFERENCE.object_schema({'unit': string, 'quote': string,
                                      'occurrence': {'type': 'integer', 'minimum': 1}})
    finding = REFERENCE.object_schema({
        'id': string, 'category': {'enum': REFERENCE.CATEGORIES, 'type': 'string'},
        'targets': {'type': 'array', 'items': target},
        'diagnostic': string, 'why_revision_needed': string,
        'meaning_to_preserve': string})
    assessment = REFERENCE.object_schema({
        'unit': string, 'technical_payload': string,
        'burden_decision': {'type': 'string',
                            'enum': ['no_burden', 'diagnose', 'uncertain']},
        'burden_reason': string,
        'finding_ids': {'type': 'array', 'items': string}})
    return REFERENCE.object_schema({
        'request': string, 'page': string,
        'assessments': {'type': 'array', 'items': assessment},
        'findings': {'type': 'array', 'items': finding}})


def validate(request, response):
    """Bind diagnoses to original UTF-8 bytes and enforce complete local ownership."""
    packet = OLD.verify_request(request)
    REFERENCE.exact_keys(response, ['request', 'page', 'assessments', 'findings'])
    require(response['request'] == request['request'] and response['page'] == packet['page'],
            'Response belongs to another request')
    require(isinstance(response['assessments'], list) and isinstance(response['findings'], list),
            'Expected assessments and findings')
    require(len(response['findings']) <= 128, 'Finding limit exceeded')
    order = {u['id']: i for i, u in enumerate(packet['units'])}
    owned = {unit: [] for unit in request['focus']}
    found = set(); bound = []
    for finding in response['findings']:
        REFERENCE.exact_keys(finding, ['id', 'category', 'targets', 'diagnostic',
                                       'why_revision_needed', 'meaning_to_preserve'])
        key = finding['id']
        require(isinstance(key, str) and key.strip() and key not in found,
                'Missing or duplicate finding ID')
        found.add(key)
        require(finding['category'] in REFERENCE.CATEGORIES, 'Unknown category')
        for field in ['diagnostic', 'why_revision_needed', 'meaning_to_preserve']:
            require(isinstance(finding[field], str) and
                    0 < len(finding[field].strip()) <= 2400, 'Missing or excessive diagnosis')
        require(isinstance(finding['targets'], list) and
                1 <= len(finding['targets']) <= 8, 'Invalid target count')
        targets = [REFERENCE.bind_target(packet, target) for target in finding['targets']]
        require(len({(t['span']['start'], t['span']['end']) for t in targets}) == len(targets),
                'Duplicate target')
        owner = min((t['unit'] for t in targets), key=order.__getitem__)
        require(owner in owned, 'Finding belongs to another focus window')
        owned[owner].append(key)
        bound.append({**copy.deepcopy(finding), 'targets': targets})
    seen = set()
    for row in response['assessments']:
        REFERENCE.exact_keys(row, ['unit', 'technical_payload', 'burden_decision',
                                   'burden_reason', 'finding_ids'])
        unit = row['unit']
        require(unit in owned and unit not in seen, 'Unknown or duplicate assessment')
        require(row['burden_decision'] in ['no_burden', 'diagnose', 'uncertain'],
                'Invalid burden decision')
        for field in ['technical_payload', 'burden_reason']:
            require(isinstance(row[field], str) and
                    0 < len(row[field].strip()) <= 1200,
                    'Missing or excessive assessment explanation')
        ids = row['finding_ids']
        require(isinstance(ids, list) and all(isinstance(i, str) for i in ids) and
                len(ids) == len(set(ids)) and set(ids) == set(owned[unit]),
                'Unlinked or invented finding')
        require((row['burden_decision'] == 'diagnose') == bool(ids),
                'Burden decision and diagnoses disagree')
        seen.add(unit)
    require(seen == set(request['focus']), 'Incomplete assessment inventory')
    return {'request': request['request'], 'page': packet['page'],
            'source_sha256': packet['source_sha256'],
            'assessments': copy.deepcopy(response['assessments']), 'findings': bound}


def assemble(page_requests, responses):
    """A page is available only after every fixed focus window validates."""
    require(bool(page_requests), 'No page requests')
    packet = OLD.verify_request(page_requests[0])
    expected = OLD.requests(packet)
    require(page_requests == expected, 'Incomplete or changed page partition')
    require(len(responses) == len(expected), 'Incomplete page results')
    assessments = []; findings = []
    for request, response in zip(expected, responses):
        result = validate(request, response)
        assessments.extend({**row, 'finding_ids': [request['request'] + '/' + key
                                                  for key in row['finding_ids']]}
                           for row in result['assessments'])
        findings.extend({**f, 'id': request['request'] + '/' + f['id']}
                        for f in result['findings'])
    return {'page': packet['page'], 'source_sha256': packet['source_sha256'],
            'coverage': [{'unit': row['unit'],
                          'status': 'uncertain' if row['burden_decision'] == 'uncertain'
                          else 'reviewed'} for row in assessments],
            'assessments': assessments, 'findings': findings}


def prepare(prior_packet, output):
    """Freeze six exposed pages; never use labels to select focus units."""
    REFERENCE.verify_freeze(prior_packet)
    source_manifest = prior_packet / 'manifest.json'
    original = read(source_manifest)
    require(len(original['pages']) == 36, 'Exposed page scope changed')
    require(digest((prior_packet / 'freeze.json').read_bytes()) ==
            read(OLD.PRIOR / 'evaluation-freeze.json')['packet_freeze_sha256'],
            'Not the original frozen experiment')
    source_references = [row['reference'] for row in original['pages']]
    require(len(set(source_references)) == 36 and
            sum(ref.startswith('whole/') for ref in source_references) == 26 and
            sum(ref.startswith('context/') for ref in source_references) == 6 and
            sum(ref.startswith('confirmation/') for ref in source_references) == 4,
            'Original page denominators changed')
    selected = [row for row in original['pages'] if row['reference'] in PAGES]
    require([row['reference'] for row in selected] == PAGES,
            'Context page selection changed')
    protocol = read(ROOT / 'protocol.json')
    output.mkdir(parents=True, exist_ok=False)
    (output / 'requests').mkdir()
    records = []
    for row in selected:
        path = prior_packet / 'requests' / (row['page'] + '.json')
        require(digest(path.read_bytes()) == row['request_sha256'], 'Original request changed')
        packet = read(path)
        require(packet['page'] == row['page'] and packet['source_sha256'] == row['source_sha256'],
                'Original source identity changed')
        for request in OLD.requests(packet):
            target = output / 'requests' / (request['request'] + '.json')
            write(target, request)
            records.append({'request': request['request'], 'page': row['page'],
                            'reference': row['reference'], 'focus_units': len(request['focus']),
                            'source_sha256': request['source_sha256'],
                            'request_sha256': digest(target.read_bytes()),
                            'request_bytes': target.stat().st_size})
    require(len(records) == protocol['max_calls'] and
            sum(r['focus_units'] for r in records) == 310, 'Pilot request scope changed')
    for name in ['diagnosis.py', 'execute.py', 'evaluate.py', 'freeze.py',
                 'protocol.json', 'prompt.md', 'DESIGN.md', 'README.md']:
        (output / name).write_bytes((ROOT / name).read_bytes())
    (output / 'source-manifest.json').write_bytes(source_manifest.read_bytes())
    write(output / 'response-schema.json', schema())
    write(output / 'manifest.json', {
        'version': 1, 'phase': 'exposed_context_pilot_preparation', 'model_calls': 0,
        'source_pages': 36, 'selected_pages': 6,
        'source_references': source_references,
        'source_manifest_sha256': digest(source_manifest.read_bytes()),
        'prior_packet_freeze_sha256': digest((prior_packet / 'freeze.json').read_bytes()),
        'accounting_code_sha256': digest((ACCOUNTING / 'accounting.py').read_bytes()),
        'source_binder_sha256': digest((OLD.PRIOR / 'study.py').read_bytes()),
        'max_focus_units': MAX_UNITS, 'requests': records})
    REFERENCE.freeze(output)
    print(json.dumps({'source_pages': 36, 'pilot_pages': 6, 'requests': len(records),
                      'focus_units': sum(r['focus_units'] for r in records),
                      'request_bytes': sum(r['request_bytes'] for r in records),
                      'model_calls': 0, 'execution_authorized': False}, indent=2))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('prior_packet', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    prepare(args.prior_packet, args.output)
