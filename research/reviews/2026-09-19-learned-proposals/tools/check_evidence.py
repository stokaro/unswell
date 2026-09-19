#!/usr/bin/env python3
"""Verify source, split, score, summary and optional Go rerun identities."""
import argparse
import gzip
import json
import math
from pathlib import Path
import subprocess
import tarfile

from assemble import AUDIT, STUDY, build_input, sha
from summarize import summarize
from rule_reference import reference


def validate_model(model, dataset, feature_limit=128):
    fold = model['evaluation_fold']
    pages = {p['id']:p for p in dataset['pages']}
    expected = {(p['id'],i) for p in pages.values() if p['fold'] == fold for i in range(len(p['units']))}
    train = [p for p in pages.values() if p['fold'] not in (fold,(fold+1)%5)]
    if model['training_groups'] != sorted({p['group'] for p in train if any(u['label'] is not None for u in p['units'])}):
        raise ValueError('Training groups changed or evaluation groups entered training')
    if model['training_rows'] != sum(u['label'] is not None for p in train for u in p['units']):
        raise ValueError('Training label count changed')
    features = model['features']
    if not 1 <= len(features) <= feature_limit or len(set(features)) != len(features):
        raise ValueError('Invalid feature contract')
    parameters = model['parameters']
    for name in ('Means','Scales','Weights'):
        if len(parameters[name]) != len(features) or not all(math.isfinite(x) for x in parameters[name]):
            raise ValueError('Invalid model parameters')
    if not math.isfinite(parameters['Intercept']) or not all(x > 0 for x in parameters['Scales']):
        raise ValueError('Invalid model scale or intercept')
    point = model['operating_point']
    development = [u for p in pages.values() if p['fold'] == (fold+1)%5 for u in p['units']]
    if point['development_positives'] != sum(u['label']==1 for u in development) or point['development_negatives'] != sum(u['label']==0 for u in development):
        raise ValueError('Development label count changed')
    if point['available']:
        tp,fp = point['development_true'],point['development_false']
        if point['threshold'] is None or not math.isfinite(point['threshold']) or tp <= 0 or fp < 0:
            raise ValueError('Invalid operating point')
        if 100*fp > point['development_negatives'] or 100*tp < 85*(tp+fp):
            raise ValueError('Development operating constraints were violated')
    elif point['threshold'] is not None or point.get('reason') != 'no_operating_point':
        raise ValueError('Unavailable operating point is not explicit')
    actual = set()
    for row in model['predictions']:
        key = row['page'],row['unit']
        if key not in expected or key in actual:
            raise ValueError('Duplicate or unexpected evaluation unit')
        actual.add(key)
        page = pages[row['page']]
        if row['label'] != page['units'][row['unit']]['label'] or row['cohort'] != page['cohort']:
            raise ValueError('Evaluation label or cohort changed')
        if not math.isfinite(row['raw_score']):
            raise ValueError('Nonfinite score')
        selected = point['available'] and row['raw_score'] >= point['threshold']
        if row['selected'] != selected:
            raise ValueError('Prediction does not use the frozen operating point')
    if actual != expected:
        raise ValueError('Missing evaluation units')


def validate_report(report, dataset):
    expected = {(kind,fold) for kind in ('L','S','W','SW') for fold in range(5)}
    actual = {(m['kind'],m['evaluation_fold']) for m in report['models']}
    if actual != expected or len(report['models']) != len(expected):
        raise ValueError('Missing or duplicate fitted model')
    if report['rows'] != sum(len(p['units']) for p in dataset['pages']):
        raise ValueError('Prepared row count changed')
    for model in report['models']:
        validate_model(model,dataset)


def check(binary=None):
    dataset,plan = build_input()
    raw = json.dumps(dataset,sort_keys=True,ensure_ascii=False).encode()+b'\n'
    plan.update(input_json_sha256=sha(raw),protocol_sha256=sha((STUDY/'protocol.md').read_bytes()))
    if plan != json.loads((STUDY/'split-plan.json').read_text()):
        raise ValueError('Split plan is not reproducible from the frozen sources')
    record = json.loads((STUDY/'measurement-record.json').read_text())
    if record['status'] != 'complete' or record['exit_code'] != 0 or record['input_json_sha256'] != sha(raw):
        raise ValueError('Incomplete or mismatched run')
    if record['protocol_sha256'] != plan['protocol_sha256'] or record['split_plan_sha256'] != sha((STUDY/'split-plan.json').read_bytes()):
        raise ValueError('Frozen protocol or plan changed')
    for name,key in [('trainer-source.tar.gz','source_snapshot_sha256'),('predictions.json.gz','predictions_sha256')]:
        if sha((STUDY/name).read_bytes()) != record[key]:
            raise ValueError('Recorded artifact changed: '+name)
    with tarfile.open(STUDY/'trainer-source.tar.gz') as archive:
        identities = {m.name:sha(archive.extractfile(m).read()) for m in archive if m.isfile()}
    if identities != record['tool_files']:
        raise ValueError('Trainer snapshot changed')
    report_bytes = gzip.decompress((STUDY/'predictions.json.gz').read_bytes())
    report = json.loads(report_bytes)
    if report['input_sha256'] != sha(raw):
        raise ValueError('Prediction input changed')
    validate_report(report,dataset)
    audit = json.loads(gzip.decompress((AUDIT/'measurements.json.gz').read_bytes()))
    summary = summarize(report,audit,plan)
    if summary != json.loads((STUDY/'summary.json').read_text()):
        raise ValueError('Summary is not derived from retained predictions')
    if reference(audit,summary) != json.loads((STUDY/'rule-reference.json').read_text()):
        raise ValueError('Prior rule credit or learned retrieval overlap changed')
    if binary:
        result = subprocess.run([str(binary.resolve())],input=raw,capture_output=True,check=True)
        if result.stdout != report_bytes:
            raise ValueError('Repeated Go fit did not reproduce the retained bytes')
    return dict(status='verified',pages=len(dataset['pages']),groups=len(plan['groups']),models=len(report['models']),
                input_json_sha256=sha(raw),predictions_sha256=record['predictions_sha256'],
                rerun_binary_sha256=sha(binary.read_bytes()) if binary else None,
                repeated_fit='byte-identical' if binary else 'not-run')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--rerun',type=Path)
    print(json.dumps(check(parser.parse_args().rerun),indent=2))
