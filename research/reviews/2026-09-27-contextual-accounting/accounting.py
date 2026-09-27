#!/usr/bin/env python3
"""Prepare bounded editorial inventories using the existing engine's raw spans."""
import argparse
import copy
import importlib.util
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
PRIOR = ROOT.parent / '2026-09-26-contextual-evidence'
SPEC = importlib.util.spec_from_file_location('source_bound_reference', PRIOR / 'study.py')
REFERENCE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(REFERENCE)
digest, read, write, require = REFERENCE.digest, REFERENCE.read, REFERENCE.write, REFERENCE.require
MAX_UNITS = 64


def source_units(packet):
    """Expose exact UTF-8 slices without parsing or normalizing the source again."""
    REFERENCE.exact_keys(packet, ['page', 'source_sha256', 'source', 'format', 'units', 'protected'])
    raw = packet['source'].encode('utf-8')
    require(digest(raw) == packet['source_sha256'], 'Source identity changed')
    require(isinstance(packet['units'], list) and packet['units'], 'No source units')
    seen = set(); units = []; previous_end = 0
    for unit in packet['units']:
        REFERENCE.exact_keys(unit, ['id', 'span'])
        REFERENCE.exact_keys(unit['span'], ['start', 'end'])
        start, end = unit['span']['start'], unit['span']['end']
        require(isinstance(unit['id'], str) and unit['id'] and unit['id'] not in seen,
                'Invalid or duplicate unit ID')
        require(type(start) is int and type(end) is int and
                previous_end <= start < end <= len(raw), 'Unordered or invalid unit range')
        units.append({**copy.deepcopy(unit), 'raw': raw[start:end].decode('utf-8')})
        seen.add(unit['id']); previous_end = end
    for region in packet['protected']:
        start, end = region['span']['start'], region['span']['end']
        require(type(start) is int and type(end) is int and 0 <= start <= end <= len(raw),
                'Invalid protected range')
        raw[start:end].decode('utf-8')
    return units


def requests(packet):
    units = source_units(packet)
    return [{**copy.deepcopy(packet), 'units': copy.deepcopy(units),
             'request': f"{packet['page']}-part-{offset // MAX_UNITS + 1:03}",
             'focus': [u['id'] for u in units[offset:offset + MAX_UNITS]]}
            for offset in range(0, len(units), MAX_UNITS)]


def verify_request(request):
    REFERENCE.exact_keys(request, ['page', 'source_sha256', 'source', 'format', 'units',
                                   'protected', 'request', 'focus'])
    packet = {key: copy.deepcopy(request[key]) for key in
              ['page', 'source_sha256', 'source', 'format', 'units', 'protected']}
    for unit in packet['units']:
        REFERENCE.exact_keys(unit, ['id', 'span', 'raw'])
        unit.pop('raw')
    expected = source_units(packet)
    require(expected == request['units'], 'Raw unit slice changed')
    require(isinstance(request['focus'], list) and
            1 <= len(request['focus']) <= MAX_UNITS, 'Invalid focus size')
    ids = [u['id'] for u in expected]
    require(request['focus'][0] in ids, 'Unknown focus unit')
    offset = ids.index(request['focus'][0])
    require(offset % MAX_UNITS == 0 and
            request['focus'] == ids[offset:offset + MAX_UNITS], 'Changed focus partition')
    require(request['request'] == f"{request['page']}-part-{offset // MAX_UNITS + 1:03}",
            'Changed request identity')
    return packet


def schema():
    original = REFERENCE.schema()
    string = {'type': 'string'}
    assessment = REFERENCE.object_schema({
        'unit': string, 'decision': {'type': 'string', 'enum': ['retain', 'revise', 'uncertain']},
        'contribution': string, 'reason': string,
        'finding_ids': {'type': 'array', 'items': string}})
    return REFERENCE.object_schema({
        'request': string, 'page': string,
        'assessments': {'type': 'array', 'items': assessment},
        'findings': original['properties']['findings']})


def validate(request, response):
    packet = verify_request(request)
    REFERENCE.exact_keys(response, ['request', 'page', 'assessments', 'findings'])
    require(response['request'] == request['request'] and response['page'] == request['page'],
            'Response belongs to another request')
    require(isinstance(response['assessments'], list), 'Missing assessments')
    # The original binder continues to enforce source ranges, protected regions,
    # exact occurrences, categories, and required diagnostic explanations.
    bound = REFERENCE.validate(packet, {
        'page': response['page'],
        'coverage': [{'unit': u['id'], 'status': 'reviewed'} for u in packet['units']],
        'findings': response['findings']})
    order = {u['id']: i for i, u in enumerate(packet['units'])}
    owned = {unit: [] for unit in request['focus']}
    for finding in bound['findings']:
        owner = min((t['unit'] for t in finding['targets']), key=order.__getitem__)
        require(owner in owned, 'Finding belongs to another focus window')
        owned[owner].append(finding['id'])
    seen = set()
    for row in response['assessments']:
        REFERENCE.exact_keys(row, ['unit', 'decision', 'contribution', 'reason', 'finding_ids'])
        unit = row['unit']
        require(unit in owned and unit not in seen, 'Unknown or duplicate assessment')
        require(row['decision'] in ['retain', 'revise', 'uncertain'], 'Invalid decision')
        for field in ['contribution', 'reason']:
            require(isinstance(row[field], str) and 0 < len(row[field].strip()) <= 1200,
                    'Missing or excessive assessment explanation')
        require(isinstance(row['finding_ids'], list) and
                all(isinstance(i, str) for i in row['finding_ids']) and
                len(set(row['finding_ids'])) == len(row['finding_ids']) and
                set(row['finding_ids']) == set(owned[unit]), 'Unlinked or invented finding')
        require((row['decision'] == 'revise') == bool(owned[unit]),
                'Revision decision and diagnostics disagree')
        seen.add(unit)
    require(seen == set(request['focus']), 'Incomplete assessment inventory')
    # Never export synthetic coverage for units outside this request's focus.
    return {'request': request['request'], 'page': packet['page'],
            'source_sha256': packet['source_sha256'],
            'assessments': copy.deepcopy(response['assessments']), 'findings': bound['findings']}


def assemble(page_requests, results):
    """A complete page requires every partition, including units with no finding."""
    require(bool(page_requests), 'No page requests')
    packet = verify_request(page_requests[0])
    expected = requests(packet)
    require(page_requests == expected, 'Incomplete or changed page partition')
    require(len(results) == len(expected), 'Incomplete page results')
    assessments = []; findings = []
    for request, response in zip(expected, results):
        result = validate(request, response)
        assessments.extend({**row, 'finding_ids': [request['request'] + '/' + key
                                                  for key in row['finding_ids']]}
                           for row in result['assessments'])
        findings.extend({**f, 'id': request['request'] + '/' + f['id']} for f in result['findings'])
    return {'page': packet['page'], 'source_sha256': packet['source_sha256'],
            'coverage': [{'unit': r['unit'],
                          'status': 'uncertain' if r['decision'] == 'uncertain' else 'reviewed'}
                         for r in assessments],
            'assessments': assessments, 'findings': findings}


def prepare(prior_packet, output):
    REFERENCE.verify_freeze(prior_packet)
    original = read(prior_packet / 'manifest.json')
    require(len(original['pages']) == 36, 'Exposed page scope changed')
    require(digest((prior_packet / 'freeze.json').read_bytes()) ==
            read(PRIOR / 'evaluation-freeze.json')['packet_freeze_sha256'],
            'Not the original frozen experiment')
    output.mkdir(parents=True, exist_ok=False); (output / 'requests').mkdir()
    records = []
    for page in original['pages']:
        path = prior_packet / 'requests' / (page['page'] + '.json')
        require(digest(path.read_bytes()) == page['request_sha256'], 'Original request changed')
        packet = read(path)
        require(packet['page'] == page['page'] and packet['source_sha256'] == page['source_sha256'],
                'Original manifest differs from request')
        for request in requests(packet):
            name = request['request']; target = output / 'requests' / (name + '.json')
            write(target, request)
            records.append({'request': name, 'page': page['page'], 'reference': page['reference'],
                            'focus_units': len(request['focus']),
                            'source_sha256': request['source_sha256'],
                            'request_sha256': digest(target.read_bytes()),
                            'request_bytes': target.stat().st_size})
    for name in ['accounting.py', 'prompt.md', 'DESIGN.md', 'execute.py',
                 'evaluate_accounting.py', 'protocol.json']:
        (output / name).write_bytes((ROOT / name).read_bytes())
    write(output / 'response-schema.json', schema())
    write(output / 'manifest.json', {
        'version': 1, 'phase': 'exposed_development_preparation', 'model_calls': 0,
        'prior_packet_freeze_sha256': digest((prior_packet / 'freeze.json').read_bytes()),
        'source_binder_sha256': digest((PRIOR / 'study.py').read_bytes()),
        'max_focus_units': MAX_UNITS, 'pages': 36, 'requests': records})
    REFERENCE.freeze(output)
    print(json.dumps({'pages': 36, 'requests': len(records),
                      'focus_units': sum(r['focus_units'] for r in records),
                      'request_bytes': sum(r['request_bytes'] for r in records),
                      'model_calls': 0, 'execution_authorized': False}, indent=2))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('prior_packet', type=Path); parser.add_argument('output', type=Path)
    args = parser.parse_args()
    prepare(args.prior_packet, args.output)
