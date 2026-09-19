#!/usr/bin/env python3
"""Derive capacity results and retain optimizer drift separately."""
import gzip
import json
from pathlib import Path
import sys

STUDY = Path(__file__).resolve().parents[1]
REFERENCE = STUDY.parent/'2026-09-19-learned-proposals'
sys.path.insert(0, str(REFERENCE/'tools'))
from summarize import summarize
from rule_reference import reference

KINDS = ('SW128','SW1024','SW8192')


def load(path):
    data = path.read_bytes()
    return json.loads(gzip.decompress(data) if path.suffix == '.gz' else data)


def compare(report, dense):
    rows = []
    for model in report['models']:
        if model['kind'] != 'SW128':
            continue
        old = next(m for m in dense['models'] if m['kind']=='SW' and m['evaluation_fold']==model['evaluation_fold'])
        if old['features'] != model['features']:
            raise ValueError('Same-capacity feature identity changed')
        prior = {(p['page'],p['unit']):p for p in old['predictions']}
        current = {(p['page'],p['unit']):p for p in model['predictions']}
        if prior.keys() != current.keys():
            raise ValueError('Comparison units changed')
        differences = {key:abs(prior[key]['raw_score']-value['raw_score']) for key,value in current.items()}
        changed = [dict(page=key[0],unit=key[1],before=prior[key]['selected'],after=value['selected'])
                   for key,value in current.items() if prior[key]['selected'] != value['selected']]
        deltas = {field:max(abs(a-b) for a,b in zip(old['parameters'][field],model['parameters'][field],strict=True))
                  for field in ('Means','Scales','Weights')}
        deltas['Intercept'] = abs(old['parameters']['Intercept']-model['parameters']['Intercept'])
        rows.append(dict(evaluation_fold=model['evaluation_fold'],features=len(model['features']),
                         max_parameter_differences=deltas,max_score_difference=max(differences.values()),
                         before_operating_point=old['operating_point'],after_operating_point=model['operating_point'],
                         changed_selections=changed))
    if len(rows) != 5:
        raise ValueError('Incomplete same-capacity comparison')
    return dict(basis='SW128 sparse versus retained SW dense; unchanged features and evaluation units',folds=rows,
                max_score_difference=max(r['max_score_difference'] for r in rows),
                changed_selections=sum(len(r['changed_selections']) for r in rows))


def derive(report):
    audit = load(STUDY.parent/'2026-09-19-representation-audit/measurements.json.gz')
    summary = summarize(report,audit,load(REFERENCE/'split-plan.json'),KINDS)
    for kind, result in summary['models'].items():
        fitted = [m for m in report['models'] if m['kind']==kind]
        result['fits'] = [dict(evaluation_fold=m['evaluation_fold'],features=len(m['features']),iterations=m['iterations'],
                              operations=m['operations'],gradient_norm=m['gradient_norm']) for m in fitted]
    return {'summary.json':summary,'rule-reference.json':reference(audit,summary),
            'optimizer-comparison.json':compare(report,load(REFERENCE/'predictions.json.gz'))}


if __name__ == '__main__':
    result = derive(load(STUDY/'predictions.json.gz'))
    for name,value in result.items():
        path = STUDY/name
        if path.exists():
            raise ValueError('Refusing to overwrite '+name)
        path.write_text(json.dumps(value,indent=2)+'\n')
    print(json.dumps({kind:{key:value[key] for key in ('true_positive','false_positive','recall','precision',
        'unlabeled_selected','event_retrieval')} for kind,value in result['summary.json']['models'].items()},indent=2))
    comparison = result['optimizer-comparison.json']
    print('Optimizer comparison:',comparison['max_score_difference'],comparison['changed_selections'])
