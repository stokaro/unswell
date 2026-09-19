#!/usr/bin/env python3
"""Verify saved supervised evidence without running or downloading a model."""
import gzip
import hashlib
import json
import math
from pathlib import Path
import sys

STUDY = Path(__file__).resolve().parents[1]
PARENT = STUDY.parent/'2026-09-19-context-encoders'
LEXICAL = STUDY.parent/'2026-09-19-learned-proposals'
sys.path.insert(0,str(LEXICAL/'tools'))
from summarize import summarize
from rule_reference import reference


def load(path):
    raw = path.read_bytes()
    return json.loads(gzip.decompress(raw) if path.suffix == '.gz' else raw)


def require(condition, message):
    if not condition:
        raise ValueError(message)


def threshold(rows):
    positive = sum(r['label'] == 1 for r in rows)
    negative = len(rows)-positive
    result = dict(available=False,threshold=None,development_positives=positive,
                  development_negatives=negative,development_true=0,development_false=0,
                  reason='no_operating_point')
    ordered = sorted(rows,key=lambda r:-r['raw_score'])
    tp = fp = i = 0
    while i < len(ordered):
        score = ordered[i]['raw_score']
        while i < len(ordered) and ordered[i]['raw_score'] == score:
            tp += ordered[i]['label'] == 1
            fp += ordered[i]['label'] == 0
            i += 1
        if tp and 100*fp <= negative and 100*tp >= 85*(tp+fp):
            if not result['available'] or tp > result['development_true'] or (
                    tp == result['development_true'] and fp < result['development_false']):
                result.update(available=True,threshold=score,development_true=tp,
                              development_false=fp)
                result.pop('reason',None)
    return result


def better(a,b):
    if not a['available']:
        return False
    return (not b['available'] or a['development_true'] > b['development_true'] or
            a['development_true'] == b['development_true'] and
            a['development_false'] < b['development_false'])


def source_metadata():
    prior = load(PARENT/'predictions.json.gz')
    pages = {p['id']:p for p in load(LEXICAL/'split-plan.json')['pages']}
    metadata = {}
    for model in prior['models']:
        if model['kind'] != 'E':
            continue
        for row in model['predictions']:
            metadata[(row['page'],row['unit'])] = dict(row,
                fold=model['evaluation_fold'],group=pages[row['page']]['group'])
    return prior,metadata


def validate(report,metadata,prior):
    require(len(report['models']) == 10 and report['rows'] == len(metadata),'Wrong denominator')
    prior_models = {m['evaluation_fold']:m for m in prior['models'] if m['kind'] == 'E'}
    parameter_names = load(STUDY/'fixture/ft-v3/report.json')['trainable_encoder_variables']
    for kind in ('H','FT'):
        models = [m for m in report['models'] if m['kind'] == kind]
        require(sorted(m['evaluation_fold'] for m in models) == list(range(5)), 'Missing fold')
        seen = set()
        for m in models:
            fold = m['evaluation_fold']
            require(m['input_sha256'] == prior['input_sha256'] and
                    m['encoder_sha256'] == prior['encoder_sha256'],'Fit identity drift')
            expected = prior_models[fold]
            require(m['training_groups'] == expected['training_groups'], 'Training group drift')
            require(m['training_rows'] == expected['training_rows'], 'Training row drift')
            require(math.isfinite(m['initial_max_score_error']) and m['initial_max_score_error'] <= 1e-4,'Warm-start mismatch')
            require(math.isfinite(m['reload_max_score_error']) and m['reload_max_score_error'] <= 1e-5,'Reload mismatch')
            require(len(m['epoch_mean_losses']) == 5 and all(math.isfinite(x) for x in m['epoch_mean_losses']),'Incomplete training')
            names = m['trainable_encoder_variables'] or []
            require(len(names) == (32 if kind == 'FT' else 0) and len(set(names)) == len(names),'Wrong trainable parameters')
            require(names == (parameter_names if kind == 'FT' else []),'Trainable identity drift')
            require(all(n.startswith(('encoder.layer.','onnx::MatMul_')) for n in names),'Graph constant trained')
            require([c['epoch'] for c in m['checkpoints']] == [0,1,3,5],'Missing checkpoint')
            devkeys = {key for key,r in metadata.items() if r['fold'] == (fold+1)%5 and
                       r['label'] is not None and r['raw_score'] is not None}
            chosen = None
            for c in m['checkpoints']:
                rows = c['development_predictions']
                keys = [(r['page'],r['unit']) for r in rows]
                require(len(keys) == len(devkeys) and set(keys) == devkeys,'Development fold drift')
                for r in rows:
                    original = metadata[(r['page'],r['unit'])]
                    require(r['label'] == original['label'] and math.isfinite(r['raw_score']),'Development label or score drift')
                require(threshold(rows) == c['operating_point'],'Threshold not reproducible')
                if chosen is None or better(c['operating_point'],chosen['operating_point']):
                    chosen = c
            require(chosen['epoch'] == m['selected_epoch'] and chosen['operating_point'] == m['operating_point'],'Checkpoint selection changed')
            point = m['operating_point']
            for r in m['predictions']:
                key = (r['page'],r['unit'])
                require(key in metadata and key not in seen,'Unknown or duplicate target')
                seen.add(key)
                original = metadata[key]
                require(original['fold'] == fold,'Evaluation fold drift')
                require(all(r[k] == original[k] for k in ('words','cohort','label')),'Target metadata drift')
                available = original['raw_score'] is not None
                require(available == (r['raw_score'] is not None),'Availability drift')
                if available:
                    require(math.isfinite(r['raw_score']),'Nonfinite evaluation score')
                else:
                    require(r.get('reason') == 'sequence_limit','Missing abstention reason')
                require(r['selected'] == (available and point['available'] and r['raw_score'] >= point['threshold']),'Selection drift')
        require(seen == set(metadata),'Dropped evaluation target')


def derive(report):
    audit = load(STUDY.parent/'2026-09-19-representation-audit/measurements.json.gz')
    result = summarize(report,audit,load(LEXICAL/'split-plan.json'),('H','FT'))
    for kind,entry in result['models'].items():
        entry['selected_epochs'] = [m['selected_epoch'] for m in report['models'] if m['kind'] == kind]
    return result,reference(audit,result)


def main():
    record = load(STUDY/'measurement-record.json')
    for name,expected in record['files'].items():
        require(hashlib.sha256((STUDY/name).read_bytes()).hexdigest() == expected,'Changed evidence: '+name)
    report = load(STUDY/'predictions.json.gz')
    prior,metadata = source_metadata()
    require(report['input_sha256'] == prior['input_sha256'] and report['encoder_sha256'] == prior['encoder_sha256'],'Source identity drift')
    validate(report,metadata,prior)
    summary,rules = derive(report)
    require(summary == load(STUDY/'summary.json') and rules == load(STUDY/'rule-reference.json'),'Derived metrics changed')
    refs = load(STUDY/'numerical-reference.json')
    require(len(refs) == 10 and all(r['passed'] and r['max_absolute_error'] <= 1e-4 for r in refs),'Independent numerical reference failed')
    fitted = {(m['kind'],m['evaluation_fold']):m for m in report['models']}
    require({(r['kind'],r['evaluation_fold']) for r in refs} == set(fitted),'Reference fold mismatch')
    for r in refs:
        model = fitted[(r['kind'],r['evaluation_fold'])]
        require(r['snapshot_sha256'] == model['snapshot_sha256'],'Reference snapshot mismatch')
        require(r['selected_epoch'] == model['selected_epoch'],'Reference epoch mismatch')
        require(r['snapshot_audit']['graph_preserved'],'Saved graph changed')
        changed = bool(r['snapshot_audit']['changed_transformer_initializers'])
        require(changed == (model['kind'] == 'FT' and model['selected_epoch'] > 0),'Unexpected encoder update')
    fixture = load(STUDY/'fixture/repeat.json')
    require(fixture['snapshot_byte_identical'] and fixture['head_byte_identical'] and
            fixture['same_numerical_report'],'Synthetic repeat failed')
    for name in ('reference.json','long-reference.json'):
        result = load(STUDY/'fixture'/name)
        require(result['passed'] and result['max_absolute_error'] <= 1e-4,'Synthetic reference failed')
    repeat = load(STUDY/'real-repeat.json')
    require(repeat['numerical_report_identical'] and all(repeat['files_byte_identical'].values()),'Real-data repeat failed')
    require(repeat['snapshot_sha256'] == fitted[('FT',0)]['snapshot_sha256'],'Repeated snapshot identity drift')
    print(json.dumps({'models':10,'evaluation_rows':2*len(metadata),'status':'research evidence verified; no diagnostic qualification'}))


if __name__ == '__main__':
    main()
