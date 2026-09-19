#!/usr/bin/env python3
"""Freeze page groups and exposed unit labels before Go-only model fitting."""
import argparse
from collections import Counter, defaultdict
import gzip
import hashlib
import json
from pathlib import Path
import sys

STUDY = Path(__file__).resolve().parents[1]
AUDIT = STUDY.parent/'2026-09-19-representation-audit'
sys.path.insert(0, str(AUDIT/'tools'))
from prepare import prepare as original_inputs
from verify import verify


def sha(data):
    return hashlib.sha256(data).hexdigest()


def group_pages(dataset, report):
    parent = {page['id']:page['id'] for page in dataset['pages']}

    def root(page):
        while parent[page] != page:
            parent[page] = parent[parent[page]]
            page = parent[page]
        return page

    seen, short = {}, defaultdict(set)
    for source in report['prepared_features']['sources']:
        for unit in source['units']:
            if unit['binding']['kind'] == 'sentence':
                continue
            words = next(v['number'] for v in unit['values'] if v['id'] == 'prose-words')
            key = unit['binding']['text_sha256']
            if words < 12:
                short[key].add(source['path'])
                continue
            if key in seen:
                left, right = sorted((root(seen[key]), root(source['path'])))
                parent[right] = left
            seen[key] = source['path']
    groups = defaultdict(list)
    for page in parent:
        groups[root(page)].append(page)
    ordered = sorted(groups, key=lambda key: sha(('review-proposals-v1\n'+key).encode()))
    if len(ordered) < 5:
        raise ValueError('Too few independent source groups')
    assignment = {page:(key, i % 5) for i,key in enumerate(ordered) for page in groups[key]}
    short_overlap = [v for v in short.values() if len(v) > 1]
    return assignment, dict(groups={k:sorted(v) for k,v in sorted(groups.items())},
                            short_reused_text_hashes=len(short_overlap),
                            short_reused_page_memberships=sum(map(len,short_overlap)))


def build_input():
    verify()
    dataset = original_inputs()
    report = json.loads(gzip.decompress((AUDIT/'representation.json.gz').read_bytes()))
    labels = json.loads(gzip.decompress((AUDIT/'supervision.json.gz').read_bytes()))
    known = {(r['page'],r['unit']):r for r in labels['rows']}
    sources = {s['path']:s for s in report['prepared_features']['sources']}
    tokens = {s['path']:s for s in report['original_tokens']}
    assignment, grouping = group_pages(dataset, report)
    pages, counts = [], defaultdict(Counter)
    for page in sorted(dataset['pages'], key=lambda p:p['id']):
        group, fold = assignment[page['id']]
        units = []
        for unit in sources[page['id']]['units']:
            if unit['binding']['kind'] == 'sentence':
                continue
            label = known.get((page['id'],len(units)),{}).get('label')
            units.append(dict(binding=unit['binding'], label=label))
            counts[fold]['unlabeled' if label is None else 'positive' if label else 'negative'] += 1
        pages.append(dict(id=page['id'], repository=page['repository'], source_group=page['repository']+'/'+page['original_path'],
                          cohort=page['cohort'], format=page['format'], sha256=page['sha256'], text=page['text'],
                          group=group, fold=fold, units=units,
                          excluded_blocks=[b['block_id'] for b in tokens[page['id']]['blocks'] if b['excluded']]))
    return dict(version=1, basis='exposed-assistant-unit-selection', pages=pages), dict(grouping,
        pages=[{k:p[k] for k in ('id','source_group','sha256','cohort','group','fold')} for p in pages],
        folds={str(k):dict(v) for k,v in sorted(counts.items())})


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    if args.output.exists():
        raise ValueError('Input output must be new')
    dataset, plan = build_input()
    raw = json.dumps(dataset, sort_keys=True, ensure_ascii=False).encode()+b'\n'
    plan['input_json_sha256'] = sha(raw)
    plan['protocol_sha256'] = sha((STUDY/'protocol.md').read_bytes())
    plan_path = STUDY/'split-plan.json'
    if plan_path.exists():
        if json.loads(plan_path.read_text()) != plan:
            raise ValueError('Reconstructed plan differs from the frozen plan')
    else:
        plan_path.write_text(json.dumps(plan,indent=2)+'\n')
    args.output.write_bytes(gzip.compress(raw, mtime=0))
    print(json.dumps(dict(groups=len(plan['groups']),folds=plan['folds'],short_overlap=plan['short_reused_text_hashes']),indent=2))
