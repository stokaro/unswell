#!/usr/bin/env python3
"""Derive exposed unit-selection labels without calling unmarked prose clean."""
from collections import Counter, defaultdict
import gzip
import json
from pathlib import Path

from evaluate import event_positions, positions
from prepare import prepare

STUDY = Path(__file__).resolve().parents[1]


def label_unit(unit_bytes, defects, uncertain, controls):
    positives = [i for i, target in enumerate(defects) if target and target <= unit_bytes]
    if positives:
        return 1, positives
    if any(target & unit_bytes for target in defects + uncertain):
        return None, []
    negatives = [i for i, target in enumerate(controls) if unit_bytes and unit_bytes <= target]
    if negatives:
        return 0, negatives
    return None, []


def derive(dataset, report):
    sources = {s['path']: s for s in report['prepared_features']['sources']}
    rows, counts, cohorts, lengths = [], Counter(), defaultdict(Counter), defaultdict(Counter)
    for page in dataset['pages']:
        units = [u for u in sources[page['id']]['units'] if u['binding']['kind'] != 'sentence']
        prose = set().union(*(positions(u['segments']) for u in units))
        targets = {kind: [event_positions(e) & prose for e in page['review'][kind]]
                   for kind in ('defects', 'uncertain', 'controls')}
        for ordinal, unit in enumerate(units):
            unit_bytes = positions(unit['segments'])
            label, events = label_unit(unit_bytes, targets['defects'], targets['uncertain'], targets['controls'])
            name = 'unlabeled' if label is None else 'positive' if label else 'negative'
            counts[name] += 1
            cohorts[page['cohort']][name] += 1
            words = next(v['number'] for v in unit['values'] if v['id'] == 'prose-words')
            band = '1-9' if words < 10 else '10-39' if words < 40 else '40+'
            lengths[band][name] += 1
            if label is None:
                continue
            kind = 'defects' if label else 'controls'
            evidence = [page['review'][kind][i].get('id', f'{kind}-{i+1}') for i in events]
            rows.append(dict(page=page['id'], unit=ordinal, source_group=page['repository']+'/'+page['original_path'],
                             source_sha256=page['sha256'], text_sha256=unit['binding']['text_sha256'],
                             context_sha256=unit['binding']['context_sha256'], kind=unit['binding']['kind'],
                             segments=unit['binding']['segments'], words=words, label=label,
                             evidence=evidence, label_basis='exposed-assistant-unit-selection'))
    return dict(version=1, target='contains-complete-reviewed-edit-target',
                basis='development-only; derived unit selection, not defect localization or calibrated quality',
                counts=dict(counts), cohorts={k:dict(v) for k,v in sorted(cohorts.items())},
                word_bands={k:dict(v) for k,v in sorted(lengths.items())}, rows=rows)


if __name__ == '__main__':
    report = json.loads(gzip.decompress((STUDY/'representation.json.gz').read_bytes()))
    result = derive(prepare(), report)
    path = STUDY/'supervision.json.gz'
    if path.exists():
        raise ValueError('Output must be new')
    path.write_bytes(gzip.compress((json.dumps(result, sort_keys=True)+'\n').encode(), mtime=0))
    print(json.dumps({k:v for k,v in result.items() if k != 'rows'}, indent=2))
