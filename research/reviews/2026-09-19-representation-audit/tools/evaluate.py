#!/usr/bin/env python3
"""Measure target representability, never semantic detection, on exposed pages."""
import argparse
from collections import Counter, defaultdict
import gzip
import json
from pathlib import Path


def positions(spans):
    result = set()
    for span in spans:
        if not 0 <= span['start'] < span['end']:
            raise ValueError('Invalid source segment')
        result.update(range(span['start'], span['end']))
    return result


def event_positions(event):
    return positions(event['targets'])


def classify(target, blocks, pieces, sentences):
    eligible = target & set().union(*blocks.values())
    touched_blocks = [key for key, value in blocks.items() if eligible & value]
    touched_pieces = [i for i, value in enumerate(pieces) if eligible & value]
    one_block = bool(eligible) and any(eligible <= b for b in blocks.values())
    one_piece = bool(eligible) and any(eligible <= p for p in pieces)
    one_sentence = bool(eligible) and any(eligible <= s for s in sentences)
    if not eligible:
        category = 'no_eligible_prose'
    elif not eligible <= set().union(*pieces):
        category = 'prepared_token_loss'
    elif not one_block:
        category = 'multiple_blocks'
    elif not one_piece:
        category = 'protected_gap_split'
    else:
        category = 'one_prepared_piece'
    return dict(eligible_bytes=len(eligible), representation=category,
                one_block=one_block, one_piece=one_piece, one_sentence=one_sentence,
                blocks=touched_blocks, prepared_pieces=touched_pieces)


def evaluate(dataset, report):
    if report['status'] != 'complete' or not report['manifest']['complete'] or report['errors']:
        raise ValueError('Incomplete analysis')
    if report.get('abstentions') or report['manifest']['skipped_rules']:
        raise ValueError('Incomplete rule evaluation')
    expected = {p['id']: p['sha256'] for p in dataset['pages']}
    documents = {d['name']: d['source_hash'] for d in report['documents']}
    if documents != expected or len(documents) != len(report['documents']):
        raise ValueError('Document identity mismatch')
    collections = {}
    for name in ('prepared_features',):
        rows = report[name]['sources']
        collections[name] = {p['path']: p for p in rows}
        if len(collections[name]) != len(rows) or {p['path']:p['source_hash'] for p in rows} != expected:
            raise ValueError('Feature source identity mismatch')
    rows = report['original_tokens']
    collections['original_tokens'] = {p['path']: p for p in rows}
    if len(collections['original_tokens']) != len(rows) or {p['path']:p['source_hash'] for p in rows} != expected:
        raise ValueError('Original token source identity mismatch')
    events, page_rows = [], []
    for page in dataset['pages']:
        row, labels = analyze_page(page, collections)
        page_rows.append(row)
        events.extend(labels)
    defects = [event for event in events if event['kind'] == 'defects']
    summary = dict(pages=len(page_rows), defects=len(defects),
                   representation=dict(Counter(d['representation'] for d in defects)),
                   one_sentence=sum(d['one_sentence'] for d in defects),
                   one_piece=sum(d['one_piece'] for d in defects),
                   one_block=sum(d['one_block'] for d in defects),
                   defects_sharing_block_with_control=sum(d['shares_block_control'] for d in defects),
                   defects_sharing_piece_with_control=sum(d['shares_piece_control'] for d in defects),
                   controls=sum(e['kind'] == 'controls' for e in events),
                   uncertain=sum(e['kind'] == 'uncertain' for e in events),
                   cohorts={}, categories={})
    for dimension, destination in [('cohort', 'cohorts'), ('category', 'categories')]:
        groups = defaultdict(list)
        for defect in defects:
            groups[defect[dimension]].append(defect)
        for name, group in sorted(groups.items()):
            summary[destination][name] = dict(defects=len(group), one_sentence=sum(e['one_sentence'] for e in group),
                one_piece=sum(e['one_piece'] for e in group), one_block=sum(e['one_block'] for e in group),
                representation=dict(Counter(e['representation'] for e in group)))
    return dict(summary=summary, pages=page_rows, events=events)


def analyze_page(page, collections):
    source = collections['original_tokens'][page['id']]
    prepared = collections['prepared_features'][page['id']]
    blocks = {u['block_id']: positions(u['segments']) for u in source['blocks'] if not u['excluded']}
    pieces = [positions(u['segments']) for u in prepared['units'] if u['binding']['kind'] != 'sentence']
    sentences = [positions(u['segments']) for u in prepared['units'] if u['binding']['kind'] == 'sentence']
    labels = []
    for kind in ('defects', 'uncertain', 'controls'):
        for ordinal, event in enumerate(page['review'].get(kind, []), 1):
            result = classify(event_positions(event), blocks, pieces, sentences)
            result.update(page=page['id'], id=event.get('id', f'{kind}-{ordinal}'), kind=kind,
                          category=event.get('category', 'technical_control'), cohort=page['cohort'])
            labels.append(result)
    controls = [e for e in labels if e['kind'] == 'controls']
    for event in labels:
        event['shares_block_control'] = any(set(event['blocks']) & set(c['blocks']) for c in controls)
        event['shares_piece_control'] = any(set(event['prepared_pieces']) & set(c['prepared_pieces']) for c in controls)
    return dict(page=page['id'], repository=page['repository'], reference=page['reference'],
                blocks=len(blocks), pieces=len(pieces), sentences=len(sentences)), labels


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inputs', type=Path, required=True)
    parser.add_argument('--report', type=Path, required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError('Output must be new')
    dataset = json.loads(gzip.decompress(args.inputs.read_bytes()))
    report = json.loads(gzip.decompress(args.report.read_bytes()))
    result = evaluate(dataset, report)
    args.output.write_text(json.dumps(result, indent=2)+'\n')
    print(json.dumps(result['summary'], indent=2))


if __name__ == '__main__':
    main()
