#!/usr/bin/env python3
"""Read-only replay of the published v30 long-sentence ablation for #421."""

import argparse
from collections import Counter, defaultdict
import hashlib
import json
from pathlib import Path


DENOMINATORS = {'whole': 57, 'context': 46, 'confirmation': 20}
REPORT = 'candidate-v1-replay/original-technical.json'
PUBLISHED = 'published-wasm-source-format-replay/original-technical.json'
EVENTS = 'review-v1/event-review.json'
OBSERVATIONS = 'follow-up-length-structure/observations-v1.json'
PROGRESS = '../quality-progress-20261008-published-discourse-stance-v30/progress.json'


def read(path):
    return json.loads(path.read_text())


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def require(condition, message):
    if not condition:
        raise ValueError(message)


def cohort(path):
    parts = path.replace('inputs/', '', 1).split('/')
    require(len(parts) == 2 and parts[0] in DENOMINATORS, 'Unknown cohort path')
    return parts[0]


def source_identity(report, manifest, source_root):
    require(len(manifest['sources']) == len(report['documents']) == 36,
            'Missing source page')
    by_path = {}
    raw_by_path = {}
    for source in manifest['sources']:
        path = source['path']
        require(path not in by_path, 'Repeated source page')
        require('/inputs/' in path, 'Source manifest path has no input identity')
        relative = path.split('/inputs/', 1)[-1]
        raw = (source_root / 'inputs' / relative).read_bytes()
        require(len(raw) == source['bytes'] and
                hashlib.sha256(raw).hexdigest() == source['source_sha256'],
                'Source hash or size changed')
        by_path[relative] = source
        raw_by_path[relative] = raw
    require(len(by_path) == 36, 'Source path join lost a page')
    for document in report['documents']:
        path = document['name'].replace('inputs/', '', 1)
        source = by_path.get(path)
        require(source is not None and document['source_hash'] == source['source_sha256']
                and document['bytes'] == source['bytes'],
                'Runtime document differs from original source')
    return by_path, raw_by_path


def finding_key(source_hash, rule, start, end, message):
    return source_hash, rule, start, end, message


def compare_published(report, published, sources, raw_sources):
    require(published['group'] == 'original' and published['profile'] == 'technical'
            and len(published['pages']) == 36, 'Published page inventory changed')
    report_by_path = {row['name'].replace('inputs/', '', 1): row
                      for row in report['documents']}
    native = Counter()
    for finding in report['findings']:
        path = finding['primary']['path'].replace('inputs/', '', 1)
        require(path in sources, 'Finding on unknown page')
        span = finding['primary']['span']
        require(0 <= span['start'] < span['end'] <= len(raw_sources[path]),
                'Finding byte span outside source')
        raw_sources[path][span['start']:span['end']].decode('utf-8')
        native[finding_key(sources[path]['source_sha256'], finding['rule_id'],
                           span['start'], span['end'], finding['message'])] += 1
    wasm = Counter()
    seen = set()
    for page in published['pages']:
        path = page['path']
        require(path in sources and path not in seen and
                page['sourceSha256'] == sources[path]['source_sha256'] ==
                report_by_path[path]['source_hash'],
                'Published page or source hash differs')
        seen.add(path)
        for finding in page['report']['findings']:
            wasm[finding_key(page['sourceSha256'], finding['ruleId'],
                             finding['start'], finding['end'],
                             finding['message'])] += 1
    require(seen == set(sources) and wasm == native and sum(native.values()) == 154,
            'Hidden, missing, or changed published delivery')
    return len(native)


def long_sentence_rows(report, observations):
    findings = {row['id']: row for row in report['findings']}
    require(len(findings) == len(report['findings']) == 154,
            'Duplicate or missing delivery ID')
    rows = observations['rows']
    require(observations['complete'] and len(rows) == 97,
            'Long-sentence observation inventory changed')
    observed = {}
    by_cohort = defaultdict(Counter)
    for row in rows:
        target = row['target']
        identifier = target['ID']
        require(identifier not in observed and identifier in findings,
                'Observation has missing or repeated finding')
        finding = findings[identifier]
        require(finding['rule_id'] == 'syntax.long-sentence' and
                finding['primary']['path'] == target['Path'] and
                finding['primary']['span'] == target['Span'] and
                target['Judgment'] in {'accepted', 'uncertain'},
                'Observation identity or grade changed')
        observed[identifier] = target['Judgment']
        by_cohort[cohort(target['Path'])][target['Judgment']] += 1
    require(set(observed) == {identifier for identifier, finding in findings.items()
                             if finding['rule_id'] == 'syntax.long-sentence'} and
            Counter(observed.values()) == {'accepted': 52, 'uncertain': 45},
            'Long-sentence grade or delivered inventory changed')
    return findings, observed, by_cohort


def event_impact(events, findings, long_ids, sources, raw_sources):
    require(len(events) == sum(DENOMINATORS.values()),
            'Original event denominator changed')
    status = defaultdict(Counter)
    sole_support = []
    mixed_support = []
    for event_id, event in events.items():
        name = event['original_event']['path']
        part = cohort(name)
        require(event_id.startswith(part + '/') and name in sources,
                'Event page or cohort changed')
        for target in event['original_event']['targets']:
            require(raw_sources[name][target['start']:target['end']].decode('utf-8') ==
                    target['quote'], 'Original event quote or byte span changed')
        state = event['after']
        require(state['status'] in {'full', 'partial', 'missed'},
                'Unknown event outcome')
        ids = state.get('finding_ids', [])
        require(len(ids) == len(set(ids)) and all(identifier in findings and
                findings[identifier]['primary']['path'].replace('inputs/', '', 1) == name
                for identifier in ids), 'Unsupported event witness')
        require(state['status'] != 'full' or ids, 'FULL event lacks witness')
        status[part][state['status']] += 1
        if state['status'] != 'full':
            continue
        affected = set(ids) & long_ids
        if affected and set(ids) <= long_ids:
            sole_support.append({'event_id': event_id,
                                 'finding_ids': sorted(affected), 'cohort': part})
        elif affected:
            mixed_support.append(event_id)
    require({part: sum(values.values()) for part, values in status.items()} == DENOMINATORS,
            'Event cohort denominators changed')
    require(not mixed_support, 'Ablation has ambiguous mixed witness support')
    return status, sole_support


def analyze(root, source_manifest, source_root):
    files = {key: root / relative for key, relative in
             [('report', REPORT), ('published', PUBLISHED), ('events', EVENTS),
              ('observations', OBSERVATIONS), ('progress', PROGRESS)]}
    manifest_path = source_manifest
    files['source_manifest'] = manifest_path
    report, published, events, observations, progress, manifest = (
        read(files[key]) for key in ('report', 'published', 'events',
                                    'observations', 'progress', 'source_manifest'))
    require(report['status'] == 'complete' and not report['errors'],
            'Runtime report incomplete')
    require(progress['scope'] == {'pages': 36, 'units': 4549, 'judgments': 804,
                                   'events': 123}, 'Published scope changed')
    sources, raw_sources = source_identity(report, manifest, source_root)
    compare_published(report, published, sources, raw_sources)
    findings, grades, by_cohort = long_sentence_rows(report, observations)
    states, sole_support = event_impact(events, findings, set(grades), sources,
                                        raw_sources)
    by_group = defaultdict(Counter)
    for row in observations['rows']:
        target = row['target']
        path = target['Path'].replace('inputs/', '', 1)
        by_group[sources[path]['repository']][target['Judgment']] += 1
    before = {}
    after = {}
    for part, denominator in DENOMINATORS.items():
        published_row = progress['per_cohort_actual_accounting']['after'][part]
        require(published_row['events'] == denominator and
                published_row['full'] == states[part]['full'] and
                published_row['deliveries'] == sum(cohort(item['primary']['path']) == part
                                                   for item in findings.values()),
                'Published cohort counts disagree with row replay')
        losses = by_cohort[part]
        require(all(published_row['grades'].get(grade, 0) >= losses[grade]
                    for grade in ('accepted', 'uncertain', 'rejected')),
                'Ablation grade count underflow')
        before[part] = {'full_events': states[part]['full'],
                        'event_denominator': denominator,
                        'deliveries': published_row['deliveries'],
                        'grades': published_row['grades']}
        remaining = {grade: published_row['grades'].get(grade, 0) - losses[grade]
                     for grade in ('accepted', 'uncertain', 'rejected')}
        remaining = {key: value for key, value in remaining.items() if value}
        after[part] = {'full_events': states[part]['full'] - sum(
                           item['cohort'] == part for item in sole_support),
                       'event_denominator': denominator,
                       'deliveries': published_row['deliveries'] - sum(losses.values()),
                       'grades': remaining}
        require(sum(remaining.values()) == after[part]['deliveries'],
                'Hidden delivery in ablation')
    require(sum(x['deliveries'] for x in before.values()) == 154 and
            sum(x['grades'].get('accepted', 0) for x in before.values()) == 94 and
            sum(x['full_events'] for x in before.values()) == 37,
            'Published overall census changed')
    return {'issue': 421, 'screen': 'whole-family long-sentence ablation',
            'distinct_repository_labels': len({row['repository'] for row in sources.values()}),
            'scope': progress['scope'],
            'inputs_sha256': {key: digest(path) for key, path in files.items()},
            'before': before, 'after': after,
            'long_sentence_grade_counts': dict(Counter(grades.values())),
            'long_sentence_by_cohort': {part: dict(by_cohort[part])
                                        for part in DENOMINATORS},
            'long_sentence_by_repository': {name: dict(counts)
                                            for name, counts in sorted(by_group.items())},
            'lost_full_event_witnesses': sole_support,
            'inference_features_used': [], 'model_calls': 0,
            'prospective_pages_used': False,
            'semantic_status': 'replay of frozen assistant judgments; no new editorial agreement',
            'product_qualified': False}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--v30-root', type=Path, required=True)
    parser.add_argument('--source-manifest', type=Path, required=True)
    parser.add_argument('--source-root', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    result = analyze(args.v30_root, args.source_manifest, args.source_root)
    args.output.write_text(json.dumps(result, indent=2, sort_keys=True) + '\n')
    print(json.dumps({'before': result['before'], 'after': result['after'],
                      'lost_full_event_witnesses': result['lost_full_event_witnesses']},
                     sort_keys=True))


if __name__ == '__main__':
    main()
