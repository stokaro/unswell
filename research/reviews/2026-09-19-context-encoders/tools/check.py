#!/usr/bin/env python3
"""Verify retained encoder evidence without loading models or making requests."""
import gzip
import hashlib
import json
from pathlib import Path
import sys

STUDY = Path(__file__).resolve().parents[1]
REFERENCE = STUDY.parent / '2026-09-19-learned-proposals'
sys.path.insert(0, str(REFERENCE / 'tools'))
from summarize import summarize, metrics
from rule_reference import reference


def load(path):
    raw = path.read_bytes()
    return json.loads(gzip.decompress(raw) if path.suffix == '.gz' else raw)


def derive(report):
    audit = load(STUDY.parent / '2026-09-19-representation-audit/measurements.json.gz')
    result = summarize(report, audit, load(REFERENCE / 'split-plan.json'), ('E', 'ES'))
    for kind, summary in result['models'].items():
        models = [m for m in report['models'] if m['kind'] == kind]
        rows = [p for m in models for p in m['predictions']]
        summary['available'] = sum(p['raw_score'] is not None for p in rows)
        summary['covered_metrics'] = metrics([p for p in rows if p['raw_score'] is not None])
        summary['fits'] = [{k:m[k] for k in ('evaluation_fold','iterations','operations',
                           'gradient_norm','training_rows','operating_point')} for m in models]
    return result, reference(audit, result)


def validate(report, vectors):
    if len(report['models']) != 10 or report['rows'] != len(vectors['rows']):
        raise ValueError('Model or row count mismatch')
    indexed = {(r['page'],r['unit']):r for r in vectors['rows']}
    if len(indexed) != len(vectors['rows']):
        raise ValueError('Duplicate encoded target')
    folds = {p['id']:p['fold'] for p in load(REFERENCE/'split-plan.json')['pages']}
    for kind in ('E','ES'):
        seen = set()
        models = [m for m in report['models'] if m['kind']==kind]
        if sorted(m['evaluation_fold'] for m in models) != list(range(5)):
            raise ValueError('Missing evaluation fold')
        for m in models:
            point = m['operating_point']
            if point['available'] and (100*point['development_false']>point['development_negatives'] or
                100*point['development_true']<85*(point['development_true']+point['development_false'])):
                raise ValueError('Invalid operating point')
            for p in m['predictions']:
                key = (p['page'],p['unit'])
                if key in seen or folds[p['page']] != m['evaluation_fold']:
                    raise ValueError('Repeated or misplaced evaluation unit')
                seen.add(key)
                vector = indexed[key]
                available = vector['values'] is not None
                if available != (p['raw_score'] is not None):
                    raise ValueError('Missingness mismatch')
                selected = available and point['available'] and p['raw_score'] >= point['threshold']
                if p['selected'] != selected:
                    raise ValueError('Selection mismatch')
                if not available and (p.get('reason') != 'sequence_limit' or vector['tokens'] <= 256):
                    raise ValueError('Unavailable output lacks a reason')
        if seen != set(indexed):
            raise ValueError('Evaluation denominator changed')


def main():
    manifest = load(STUDY/'measurement-record.json')
    for name, expected in manifest['files'].items():
        if hashlib.sha256((STUDY/name).read_bytes()).hexdigest() != expected:
            raise ValueError('Changed evidence: '+name)
    plan = load(STUDY/'split-plan.json')
    if hashlib.sha256((STUDY/'protocol.md').read_bytes()).hexdigest() != plan['protocol_sha256']:
        raise ValueError('Protocol changed')
    report, vectors = load(STUDY/'predictions.json.gz'), load(STUDY/'embeddings.json.gz')
    if report['encoder_sha256'] != plan['encoder_manifest_sha256'] or report['input_sha256'] != plan['input_json_sha256']:
        raise ValueError('Execution identity changed')
    if hashlib.sha256(gzip.decompress((STUDY/'embeddings.json.gz').read_bytes())).hexdigest() != report['embeddings_sha256']:
        raise ValueError('Fitted sidecar changed')
    parity = load(STUDY/'batch-reference.json')
    if not parity['passed'] or parity['max_absolute_error'] > 1e-4:
        raise ValueError('Encoder parity failed')
    validate(report,vectors)
    summary, rules = derive(report)
    if summary != load(STUDY/'summary.json') or rules != load(STUDY/'rule-reference.json'):
        raise ValueError('Derived metrics changed')
    print(json.dumps({'models':len(report['models']),'rows_per_model':report['rows'],
                      'status':'retained evidence verified; no product qualification'}))


if __name__ == '__main__':
    main()
