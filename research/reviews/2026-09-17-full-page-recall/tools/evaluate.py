#!/usr/bin/env python3
"""Validate frozen whole-page evidence and compute recall without rerunning NLP."""

import argparse
from collections import Counter
import gzip
import hashlib
import json
from pathlib import Path, PurePosixPath
import random
import tarfile

CATEGORIES = {'empty_framing', 'needless_repetition', 'wordiness',
              'unjustified_intensifiers', 'vague_claims', 'formulaic_transitions',
              'needless_complexity'}
SEMANTICS = {
    'syntax.long-sentence': {'needless_complexity'},
    'filler.wordy-phrase': {'wordiness'},
    'filler.document-metadiscourse': {'empty_framing'},
    'filler.document-justification': {'empty_framing'},
    'filler.evaluative-closure': {'empty_framing'},
    'syntax.slogan-contrast': {'empty_framing', 'vague_claims'},
    **{f'repetition.{k}': {'needless_repetition'} for k in
       ['sentence-openers', 'paragraph-openers', 'exact-sentence', 'near-sentence', 'definition-echo']},
}


def require(condition, message):
    if not condition:
        raise ValueError(message)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def encode(value):
    return json.dumps(value, sort_keys=True, separators=(',', ':'), ensure_ascii=False, allow_nan=False).encode()


def read(path):
    def pairs(items):
        result = {}
        for key, value in items:
            require(key not in result, 'Duplicate JSON key')
            result[key] = value
        return result
    def invalid(value):
        raise ValueError('Nonfinite JSON value: ' + value)
    data = path.read_bytes()
    if path.suffix == '.gz':
        data = gzip.decompress(data)
    return json.loads(data, object_pairs_hook=pairs, parse_constant=invalid)


def write(path, value):
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2, allow_nan=False) + '\n', encoding='utf-8')


def unique(rows, key, label):
    result = {r[key]: r for r in rows}
    require(len(result) == len(rows), 'Duplicate ' + label)
    return result


def overlap(a, b):
    return max(a['start'], b['start']) < min(a['end'], b['end'])


def span(source, target):
    start, end = target['start'], target['end']
    require(type(start) is int and type(end) is int and 0 <= start < end <= len(source), 'Invalid source span')
    try:
        excerpt = source[start:end].decode('utf-8')
        source[:start].decode('utf-8')
    except UnicodeDecodeError as error:
        raise ValueError('Source span splits UTF-8') from error
    if 'quote' in target:
        require(excerpt == target['quote'], 'Source quote differs')


def archive_sources(path):
    result = {}
    with tarfile.open(path) as archive:
        for info in archive:
            name = PurePosixPath(info.name)
            require(info.isfile() and not name.is_absolute() and '..' not in name.parts
                    and str(name) == info.name and info.size <= 4 * 1024 * 1024,
                    'Unsafe source archive member')
            require(info.name not in result, 'Duplicate source archive member')
            result[info.name] = archive.extractfile(info).read()
    return result


def load(root):
    freeze = read(root / 'freeze.json')
    require(freeze['version'] == 'unswell-full-page-freeze-v1'
            and freeze['phase'] == 'before_new_detector_output_review', 'Invalid freeze')
    require(set(freeze['files']) == {'protocol.md', 'manifest.json', 'annotations.json', 'inputs.tar.gz'}, 'Incomplete freeze')
    for name, digest in freeze['files'].items():
        require(sha((root / name).read_bytes()) == digest, 'Frozen artifact changed: ' + name)
    manifest = read(root / 'manifest.json')
    require(manifest['version'] == 'unswell-full-page-inputs-v1', 'Unknown input schema')
    pages = unique(manifest['pages'], 'id', 'page ID')
    unique(manifest['pages'], 'path', 'page path')
    files = archive_sources(root / 'inputs.tar.gz')
    annotations = unique(read(root / 'annotations.json'), 'page', 'page annotation')
    require(set(annotations) == set(pages), 'Incomplete page annotations')
    defects = {}
    for pid, page in pages.items():
        source = files[page['path']]
        require(sha(source) == page['sha256'] and len(source) == page['bytes'], 'Source drift')
        for notice in page['retained_notices']:
            require(sha(files[notice['path']]) == notice['sha256'], 'Notice drift')
        a = annotations[pid]
        require(a['version'] == 'unswell-page-review-v1' and a['source_sha256'] == page['sha256']
                and a['phase'] == freeze['phase'] and a['reviewer'], 'Invalid page review')
        end, last_line = 0, 0
        lines = source.splitlines(keepends=True)
        for section in a['coverage']:
            span(source, section)
            require(section['start'] == end and section['first_line'] == last_line + 1
                    and section['rationale'], 'Incomplete coverage')
            end, last_line = section['end'], section['last_line']
            require(0 < last_line <= len(lines) and len(b''.join(lines[:last_line])) == end,
                    'Coverage line/byte mismatch')
        require(end == len(source), 'Incomplete coverage')
        for kind in ['defects', 'uncertain', 'controls']:
            for event in a[kind]:
                require(event['targets'] and event['rationale'], 'Empty annotation')
                for target in event['targets']:
                    span(source, target)
                if kind != 'controls':
                    require(event['category'] in CATEGORIES and event['proposed_edit'], 'Invalid editorial event')
                    require(event['id'] not in defects, 'Duplicate editorial event')
                    defects[event['id']] = dict(event, page=pid, kind=kind)
    addenda = read(root / 'addendum.json')
    require(addenda['phase'] == 'after_detector_output_review', 'Unmarked post-freeze evidence')
    for event in addenda['events']:
        require(event['id'] not in defects and event['page'] in pages, 'Duplicate/unknown addendum event')
        require(event['category'] in CATEGORIES and event['rationale'] and event['proposed_edit']
                and event['targets'], 'Invalid addendum event')
        for target in event['targets']:
            span(files[pages[event['page']]['path']], target)
        defects[event['id']] = dict(event, kind='addendum')
    return manifest, pages, files, annotations, defects


def locations(finding):
    return [finding['primary'], *finding['related']]


def matches(finding, event, pages):
    hits = [any(loc['path'] == pages[event['page']]['path'] and overlap(segment, target)
                for loc in locations(finding) for segment in (loc.get('segments') or [loc['span']]))
            for target in event['targets']]
    # Repetition evidence must include every annotated occurrence, not merely
    # one sentence that happens to belong to a multi-location defect.
    return all(hits) if event['category'] == 'needless_repetition' else any(hits)


def validate_run(report, rows, pages, files, defects, engine_commit):
    require(report['schema_version'] == '1.0.0-alpha.1', 'Unsupported report schema')
    require(report['status'] == 'complete' and report['manifest']['complete'] and not report['errors'], 'Incomplete run')
    require(report['manifest']['tool_commit'] == engine_commit, 'Engine identity differs')
    require(not report['manifest']['skipped_rules'], 'Unexpected rule abstention')
    require(report['manifest']['include_source'] and not report['manifest']['no_gate'], 'Altered measurement policy')
    docs = unique(report['documents'], 'name', 'reported document')
    require(set(docs) == {p['path'] for p in pages.values()}, 'Missing/extra reported document')
    for page in pages.values():
        d = docs[page['path']]
        require(d['source'].encode() == files[page['path']] and d['source_hash'] == page['sha256']
                and d['bytes'] == page['bytes'] and d['format'] == page['format'], 'Report source differs')
    require(len(rows) == len(report['findings']), 'Missing finding disposition')
    index = unique(rows, 'index', 'finding disposition')
    require(set(index) == set(range(len(report['findings']))), 'Disposition index mismatch')
    for i, finding in enumerate(report['findings']):
        row = index[i]
        require(row['finding_sha256'] == sha(encode(finding)), 'Finding identity differs')
        require(row['status'] in ['actionable', 'nonactionable', 'uncertain'] and row['rationale'], 'Invalid disposition')
        require(len(row['events']) == len(set(row['events'])), 'Duplicate recall credit')
        require(bool(row['events']) == (row['status'] == 'actionable'), 'Actionable finding needs event')
        for loc in locations(finding):
            require(loc['path'] in docs, 'Unknown finding source')
            span(files[loc['path']], loc['span'])
            for segment in loc.get('segments', []):
                span(files[loc['path']], segment)
                require(loc['span']['start'] <= segment['start'] < segment['end'] <= loc['span']['end'], 'Segment outside location')
        for eid in row['events']:
            require(eid in defects, 'Unknown credited event')
            event = defects[eid]
            require(event['kind'] != 'uncertain', 'Uncertain event counted as positive')
            require(event['category'] in SEMANTICS.get(finding['rule_id'], set()), 'Semantic mismatch')
            require(matches(finding, event, pages), 'Credited event has no source overlap')
    for assessment in report['assessments']:
        require(assessment['path'] in docs, 'Unknown assessment source')
        span(files[assessment['path']], assessment['span'])
    return index


def ratio(a, b):
    return a / b if b else None


def uncertainty(rows):
    """Resample paired pages within their fixed format/length cells."""
    cells = {}
    for row in rows:
        cells.setdefault((row['format'], row['length_stratum']), []).append(row)
    rng = random.Random(9172026)
    values = []
    for _ in range(2000):
        sampled = [rng.choice(cell) for cell in cells.values() for _ in cell]
        value = ratio(sum(r['detected'] for r in sampled), sum(r['defects'] for r in sampled))
        if value is not None:
            values.append(value)
    values.sort()
    leave = [ratio(sum(x['detected'] for x in rows if x != r),
                   sum(x['defects'] for x in rows if x != r)) for r in rows]
    leave = [x for x in leave if x is not None]
    return dict(page_bootstrap_percentile_95=[values[int((len(values)-1)*q)] for q in (.025, .975)] if values else None,
                valid_bootstrap_replicates=len(values), leave_one_page_out_range=[min(leave), max(leave)] if leave else None)


def evaluate(root):
    manifest, pages, files, annotations, defects = load(root)
    reviews = read(root / 'dispositions.json')
    runs = read(root / 'runs.json')
    result = dict(version='unswell-full-page-results-v1', review='Single assistant, auxiliary development labels',
                  freeze_sha256=sha((root / 'freeze.json').read_bytes()), profiles={})
    require(set(reviews) == set(runs) == {'technical', 'strict'}, 'Missing profile')
    for profile in ['technical', 'strict']:
        record = runs[profile]
        require(record['path'] == profile + '.json.gz', 'Invalid report path')
        require(sha((root / record['path']).read_bytes()) == record['sha256'], 'Report artifact drift')
        report = read(root / record['path'])
        config_name = profile + '.yaml'
        require(sha((root / config_name).read_bytes()) == record['config_sha256'], 'Config artifact drift')
        require(report['manifest']['config_sources'] == [dict(path=config_name, kind='config',
                sha256=record['config_identity_sha256'])], 'Unexpected config provenance')
        require(record['exit_code'] in [0, 1] and record['exit_code'] == (0 if report['gate']['passed'] else 1), 'Exit/gate mismatch')
        require(report['manifest']['scoring_profile'] == profile+'-v1', 'Profile differs')
        index = validate_run(report, reviews[profile], pages, files, defects, manifest['engine_commit'])
        detected = {eid for row in index.values() for eid in row['events'] if defects[eid]['kind'] == 'defects'}
        per_page = []
        for pid, page in pages.items():
            events = annotations[pid]['defects']
            findings = [(i, f) for i, f in enumerate(report['findings']) if f['primary']['path'] == page['path']]
            document = next(d for d in report['documents'] if d['name'] == page['path'])
            counts = Counter(index[i]['status'] for i, _ in findings)
            per_page.append(dict(page=pid, cohort=page['cohort'], selection=page['selection'], format=page['format'],
                                 length_stratum=page['length_stratum'], path=page['original_path'],
                                 words=document['prose_words'], defects=len(events), detected=sum(e['id'] in detected for e in events),
                                 missed=[e['id'] for e in events if e['id'] not in detected], findings=len(findings),
                                 dispositions={k:counts[k] for k in ['actionable','nonactionable','uncertain']}))
        groups = {}
        for cohort, selection in [('ptah','sample'),('historical','sample'),('ptah','exposed_anchor')]:
            rows = [r for r in per_page if r['cohort'] == cohort and r['selection'] == selection]
            total, found = sum(r['defects'] for r in rows), sum(r['detected'] for r in rows)
            counts = Counter()
            for r in rows:
                counts.update(r['dispositions'])
            n = sum(r['findings'] for r in rows)
            categories = {}
            for category in sorted(CATEGORIES):
                ev = [e for e in defects.values() if e['kind']=='defects' and e['category']==category
                      and pages[e['page']]['cohort']==cohort and pages[e['page']]['selection']==selection]
                categories[category] = dict(defects=len(ev), detected=sum(e['id'] in detected for e in ev))
            groups[cohort+'-'+selection] = dict(pages=len(rows), defects=total, detected=found, recall=ratio(found,total),
                findings=n, dispositions=dict(counts), actionable_fraction=ratio(counts['actionable'],n),
                uncertain_fraction=ratio(counts['uncertain'],n), findings_per_1000_words=ratio(1000*n,sum(r['words'] for r in rows)),
                categories=categories, sensitivity=uncertainty(rows) if selection=='sample' else None)
        # Provisional clean units exclude frozen positives, uncertain labels,
        # post-freeze positives, and unresolved diagnostic locations.
        clean = Counter()
        for a in report['assessments']:
            if a['scope'] != 'paragraph' or a['words'] == 0:
                continue
            pid = next(pid for pid,p in pages.items() if p['path']==a['path'])
            if pages[pid]['selection'] != 'sample':
                continue
            if any(e['page']==pid and any(overlap(a['span'], t) for t in e['targets']) for e in defects.values()):
                continue
            relevant = [index[i] for i,f in enumerate(report['findings'])
                        if any(l['path']==a['path'] and overlap(a['span'],l['span']) for l in locations(f))]
            if any(row['status']=='uncertain' for row in relevant):
                continue
            key = pages[pid]['cohort']
            for minimum in [1, 20, 40]:
                if a['words'] < minimum:
                    continue
                prefix = key + '_min' + str(minimum)
                clean[prefix+'_units'] += 1
                clean[prefix+'_alarmed'] += any(row['status']=='nonactionable' for row in relevant)
        result['profiles'][profile] = dict(groups=groups, pages=per_page,
            provisional_clean_paragraphs=dict(clean), gate=report['gate'],
            exclusions=dict(Counter(e['reason'] for d in report['documents'] for e in d['excluded'])),
            abstentions=report['manifest']['skipped_rules'])
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('record', type=Path)
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    result = evaluate(args.record)
    if args.output:
        write(args.output, result)
    else:
        print(json.dumps({p:r['groups'] for p,r in result['profiles'].items()}, indent=2))


if __name__ == '__main__':
    main()
