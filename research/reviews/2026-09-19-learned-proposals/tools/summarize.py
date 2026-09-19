#!/usr/bin/env python3
"""Summarize stored proposals without granting them diagnostic credit."""
from collections import defaultdict
import gzip
import json
from pathlib import Path

STUDY = Path(__file__).resolve().parents[1]
AUDIT = STUDY.parent/'2026-09-19-representation-audit'


def fraction(numerator, denominator):
    return numerator/denominator if denominator else None


def metrics(predictions):
    positive = sum(p['label'] == 1 for p in predictions)
    negative = sum(p['label'] == 0 for p in predictions)
    tp = sum(p['selected'] and p['label'] == 1 for p in predictions)
    fp = sum(p['selected'] and p['label'] == 0 for p in predictions)
    return dict(positive=positive, negative=negative, true_positive=tp, false_positive=fp,
                precision=fraction(tp,tp+fp), recall=fraction(tp,positive),
                explicit_control_fpr=fraction(fp,negative),
                unlabeled=sum(p['label'] is None for p in predictions),
                unlabeled_selected=sum(p['selected'] and p['label'] is None for p in predictions),
                total_selected=sum(p['selected'] for p in predictions))


def grouped(rows, field):
    groups = defaultdict(list)
    for row in rows:
        groups[field(row)].append(row)
    return {str(key):metrics(value) for key,value in sorted(groups.items())}


def retrieval(events, selected):
    records = []
    for event in events:
        if event['kind'] != 'defects':
            continue
        targets = event['prepared_pieces']
        covered = bool(targets) and all((event['page'],i) in selected for i in targets)
        records.append(dict(page=event['page'],id=event['id'],category=event['category'],
                            cohort=event['cohort'],retrieved=covered))
    return records


def retrieval_counts(events):
    return dict(defects=len(events),retrieved=sum(e['retrieved'] for e in events),
                coverage=fraction(sum(e['retrieved'] for e in events),len(events)))


def summarize(report, audit, plan, kinds=('L','S','W','SW')):
    folds = {page['id']:page['fold'] for page in plan['pages']}
    models = {}
    for kind in kinds:
        fitted = [m for m in report['models'] if m['kind'] == kind]
        predictions = [p for m in fitted for p in m['predictions']]
        selected = {(p['page'],p['unit']) for p in predictions if p['selected']}
        events = retrieval(audit['events'],selected)
        bands = lambda p: '1-9' if p['words'] < 10 else '10-39' if p['words'] < 40 else '40+'
        cohorts = sorted({e['cohort'] for e in events})
        categories = sorted({e['category'] for e in events})
        models[kind] = dict(metrics(predictions),
            available_folds=sum(m['operating_point']['available'] for m in fitted),
            cohorts=grouped(predictions,lambda p:p['cohort']), word_bands=grouped(predictions,bands),
            folds=grouped(predictions,lambda p:folds[p['page']]),
            event_retrieval=retrieval_counts(events),
            event_cohorts={c:retrieval_counts([e for e in events if e['cohort']==c]) for c in cohorts},
            event_categories={c:retrieval_counts([e for e in events if e['category']==c]) for c in categories},
            retrieved_events=[{'page':e['page'],'id':e['id']} for e in events if e['retrieved']])
    return dict(version=1,basis='exposed assistant development; unit retrieval is not diagnostic recall',models=models)


if __name__ == '__main__':
    result = summarize(json.loads(gzip.decompress((STUDY/'predictions.json.gz').read_bytes())),
                       json.loads(gzip.decompress((AUDIT/'measurements.json.gz').read_bytes())),
                       json.loads((STUDY/'split-plan.json').read_text()))
    path = STUDY/'summary.json'
    if path.exists():
        raise ValueError('Summary already exists')
    path.write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({k:{key:value for key,value in v.items() if key not in
        ('cohorts','word_bands','folds','event_cohorts','event_categories','retrieved_events')}
        for k,v in result['models'].items()},indent=2))
