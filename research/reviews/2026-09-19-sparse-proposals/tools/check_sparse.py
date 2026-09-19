#!/usr/bin/env python3
"""Verify the frozen sparse run, capacity comparison and optional exact rerun."""
import argparse
import gzip
import json
import math
from pathlib import Path
import subprocess
import tarfile

from results import KINDS, REFERENCE, STUDY, derive, load
from assemble import build_input, sha
from check_evidence import validate_model


def validate_sparse(model, dataset):
    if model['kind'] not in KINDS:
        raise ValueError('Unknown capacity')
    validate_model(model,dataset,int(model['kind'][2:]))
    expected = dict(L2=0.01,Tolerance=1e-8,MaxIterations=500,MaxOperations=3000000000)
    if model['options'] != expected:
        raise ValueError('Optimizer options changed')
    if not 0 <= model['iterations'] < 500 or not 0 < model['operations'] <= 3000000000:
        raise ValueError('Optimizer budget or convergence limit changed')
    gradient = model['gradient_norm']
    if not math.isfinite(gradient) or not 0 <= gradient <= 1e-8:
        raise ValueError('Unconverged model')


def validate_report(report, dataset):
    if report['version'] != 2 or report['model_algorithm'] != 'unswell-research-sparse-logistic-lbfgs-v1':
        raise ValueError('Unexpected sparse algorithm')
    expected = {(kind,fold) for kind in KINDS for fold in range(5)}
    actual = {(m['kind'],m['evaluation_fold']) for m in report['models']}
    if actual != expected or len(report['models']) != len(expected):
        raise ValueError('Missing or duplicate fitted model')
    if report['rows'] != sum(len(p['units']) for p in dataset['pages']):
        raise ValueError('Prepared row count changed')
    for model in report['models']:
        validate_sparse(model,dataset)


def check(binary=None):
    dataset, original_plan = build_input()
    raw = json.dumps(dataset,sort_keys=True,ensure_ascii=False).encode()+b'\n'
    original_plan.update(input_json_sha256=sha(raw),protocol_sha256=sha((REFERENCE/'protocol.md').read_bytes()))
    if original_plan != load(REFERENCE/'split-plan.json'):
        raise ValueError('Source grouping or labels changed')
    plan = load(STUDY/'split-plan.json')
    if plan != dict(reference='../2026-09-19-learned-proposals/split-plan.json',
                    reference_sha256=sha((REFERENCE/'split-plan.json').read_bytes()),input_json_sha256=sha(raw),
                    protocol_sha256=sha((STUDY/'protocol.md').read_bytes())):
        raise ValueError('Frozen plan changed')
    record = load(STUDY/'measurement-record.json')
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
    dense = load(REFERENCE/'predictions.json.gz')
    for field in ('unit_contract','feature_contract','lexical_contract','provider'):
        if report[field] != dense[field]:
            raise ValueError('Preparation contract changed: '+field)
    validate_report(report,dataset)
    for name,value in derive(report).items():
        if value != load(STUDY/name):
            raise ValueError('Derived result changed: '+name)
    if binary:
        result = subprocess.run([str(binary.resolve())],input=raw,capture_output=True,check=True)
        if result.stdout != report_bytes:
            raise ValueError('Repeated Go fit did not reproduce the retained bytes')
    return dict(status='verified',pages=len(dataset['pages']),groups=len(original_plan['groups']),models=len(report['models']),
                input_json_sha256=sha(raw),predictions_sha256=record['predictions_sha256'],
                rerun_binary_sha256=sha(binary.read_bytes()) if binary else None,
                repeated_fit='byte-identical' if binary else 'not-run')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--rerun',type=Path)
    print(json.dumps(check(parser.parse_args().rerun),indent=2))
